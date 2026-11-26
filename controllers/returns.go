package controllers
import (
 "errors"
 "strings"
 "time"
 "math/big"
 "github.com/gin-gonic/gin"
 "gorm.io/gorm"
 "gorm.io/gorm/clause"
 "github.com/SamyakNaranje/E-Commerce-Application/models"
 "github.com/SamyakNaranje/E-Commerce-Application/services"
)
type Returns struct{DB *gorm.DB}
var errReturn=errors.New("return not eligible or invalid state")
func refundAmount(line,subtotal,discount int64)int64{
 if line<0 || subtotal<=0 || discount<0 || discount>subtotal || line>subtotal{return 0}
 n:=new(big.Int).Mul(big.NewInt(line),big.NewInt(subtotal-discount))
 return n.Quo(n,big.NewInt(subtotal)).Int64()
}
func(h Returns) Request(c *gin.Context){
 var in struct{ItemID uint64 `json:"order_item_id" binding:"required"`;Reason string `json:"reason" binding:"required,max=500"`}
 if c.ShouldBindJSON(&in)!=nil || strings.TrimSpace(in.Reason)==""{c.JSON(400,gin.H{"error":"invalid request"});return}
 var ret models.ReturnRequest
 err:=h.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB)error{
 var item models.OrderItem
 if e:=tx.First(&item,in.ItemID).Error;e!=nil{return e}
 var o models.Order
 if e:=tx.Clauses(clause.Locking{Strength:"UPDATE"}).Where("id = ? AND user_id = ?",item.OrderID,shoppingUser(c)).First(&o).Error;e!=nil{return e}
 if o.Status!="delivered"{return errReturn}
 var event models.OrderEvent
 if e:=tx.Where("order_id = ? AND to_status = ?",o.ID,"delivered").Order("id DESC").First(&event).Error;e!=nil{return e}
 if time.Now().After(event.CreatedAt.Add(7*24*time.Hour)){return errReturn}
 var count int64
 if e:=tx.Model(&models.ReturnRequest{}).Where("order_item_id = ?",item.ID).Count(&count).Error;e!=nil{return e};if count>0{return errReturn}
 ret=models.ReturnRequest{OrderItemID:item.ID,Reason:strings.TrimSpace(in.Reason),Status:"pending"}
 return tx.Create(&ret).Error
 })
 if errors.Is(err,errReturn){c.JSON(409,gin.H{"error":"return not eligible"});return};if err!=nil{catalogError(c,err);return};c.JSON(201,ret)
}
func(h Returns) list(c *gin.Context,admin bool){
 rows:=[]models.ReturnRequest{}
 q:=h.DB.Model(&models.ReturnRequest{})
 if !admin{q=q.Where("order_item_id IN (?)",h.DB.Table("order_items").Select("order_items.id").Joins("JOIN orders ON orders.id = order_items.order_id").Where("orders.user_id = ?",shoppingUser(c)))}
 if e:=q.Preload("OrderItem").Order("id DESC").Limit(200).Find(&rows).Error;e!=nil{catalogError(c,e);return}
 c.JSON(200,rows)
}
func(h Returns) List(c *gin.Context){h.list(c,false)}
func(h Returns) AdminList(c *gin.Context){h.list(c,true)}
func(h Returns) Update(c *gin.Context){
 var in struct{Status string `json:"status" binding:"required,oneof=approved rejected received refunded"`;Note string `json:"note" binding:"required,max=500"`;Restock bool `json:"restock"`;Reference string `json:"reference" binding:"max=120"`}
 if c.ShouldBindJSON(&in)!=nil || strings.TrimSpace(in.Note)==""{c.JSON(400,gin.H{"error":"invalid update"});return}
 err:=h.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB)error{
 var ret models.ReturnRequest
 if e:=tx.Clauses(clause.Locking{Strength:"UPDATE"}).First(&ret,"id = ?",c.Param("id")).Error;e!=nil{return e}
 if ret.Status==in.Status{return nil}
 switch in.Status{
 case "approved","rejected":if ret.Status!="pending"{return errReturn}
 case "received":
 if ret.Status!="approved"{return errReturn}
 var item models.OrderItem
 if e:=tx.First(&item,ret.OrderItemID).Error;e!=nil{return e}
 var o models.Order
 if e:=tx.First(&o,item.OrderID).Error;e!=nil{return e}
 var p models.Payment
 if e:=tx.Where("order_id = ?",o.ID).First(&p).Error;e!=nil{return e}
 if p.Status!="paid"{return errReturn}
 if in.Restock{if _,e:=services.AdjustTx(tx,item.ProductID,item.Quantity,"return received");e!=nil{return e}}
 amount:=refundAmount(item.Quantity*item.UnitPricePaise,o.SubtotalPaise,o.DiscountPaise)
 if e:=tx.Create(&models.Refund{ReturnRequestID:ret.ID,AmountPaise:amount,Status:"pending"}).Error;e!=nil{return e}
 case "refunded":
 if ret.Status!="received" || strings.TrimSpace(in.Reference)==""{return errReturn}
 result:=tx.Model(&models.Refund{}).Where("return_request_id = ? AND status = ?",ret.ID,"pending").Updates(map[string]interface{}{"status":"completed","reference":strings.TrimSpace(in.Reference)})
 if result.Error!=nil{return result.Error};if result.RowsAffected!=1{return errReturn}
 }
 if e:=tx.Model(&ret).Updates(map[string]interface{}{"status":in.Status,"admin_note":in.Note}).Error;e!=nil{return e}
 return tx.Create(&models.AuditLog{UserID:shoppingUser(c),Action:"return_"+in.Status,Resource:c.Param("id")}).Error
 })
 if errors.Is(err,errReturn){c.JSON(409,gin.H{"error":"invalid return transition"});return};if err!=nil{catalogError(c,err);return};c.Status(204)
}
func(h Returns) Refunds(c *gin.Context){
 rows:=[]models.Refund{}
 q:=h.DB.Model(&models.Refund{})
 if c.MustGet("user").(models.User).Role!="admin"{q=q.Where("return_request_id IN (?)",h.DB.Table("return_requests").Select("return_requests.id").Joins("JOIN order_items ON order_items.id = return_requests.order_item_id").Joins("JOIN orders ON orders.id = order_items.order_id").Where("orders.user_id = ?",shoppingUser(c)))}
 if e:=q.Order("id DESC").Limit(200).Find(&rows).Error;e!=nil{catalogError(c,e);return};c.JSON(200,rows)
}
