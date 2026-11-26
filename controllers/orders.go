package controllers

import (
 "errors"
 "fmt"
 "strconv"
 "github.com/gin-gonic/gin"
 "gorm.io/gorm"
 "gorm.io/gorm/clause"
 "github.com/SamyakNaranje/E-Commerce-Application/models"
 "github.com/SamyakNaranje/E-Commerce-Application/services"
)
type Orders struct{DB *gorm.DB}
var errOrderState=errors.New("invalid order transition")
func (h Orders) list(c *gin.Context,admin bool){
 page,e:=strconv.Atoi(c.DefaultQuery("page","1"))
 if e!=nil || page<1 || page>100000{c.JSON(400,gin.H{"error":"invalid page"});return}
 q:=h.DB.WithContext(c.Request.Context()).Model(&models.Order{})
 if !admin{q=q.Where("user_id = ?",shoppingUser(c))}
 if s:=c.Query("status");s!=""{q=q.Where("status = ?",s)}
 var total int64
 if e:=q.Count(&total).Error;e!=nil{catalogError(c,e);return}
 rows:=[]models.Order{}
 if e:=q.Preload("Items").Order("id DESC").Limit(20).Offset((page-1)*20).Find(&rows).Error;e!=nil{catalogError(c,e);return}
 c.JSON(200,gin.H{"items":rows,"total":total,"page":page})
}
func(h Orders) List(c *gin.Context){h.list(c,false)}
func(h Orders) AdminList(c *gin.Context){h.list(c,true)}
func(h Orders) detail(c *gin.Context,admin bool){
 var o models.Order
 q:=h.DB.WithContext(c.Request.Context()).Preload("Items").Where("id = ?",c.Param("id"))
 if !admin{q=q.Where("user_id = ?",shoppingUser(c))}
 if e:=q.First(&o).Error;e!=nil{catalogError(c,e);return}
 var p models.Payment
 if e:=h.DB.Where("order_id = ?",o.ID).First(&p).Error;e!=nil{catalogError(c,e);return}
 history:=[]models.OrderEvent{}
 if e:=h.DB.Where("order_id = ?",o.ID).Order("id ASC").Find(&history).Error;e!=nil{catalogError(c,e);return}
 c.JSON(200,gin.H{"order":o,"payment":p,"history":history})
}
func(h Orders) Detail(c *gin.Context){h.detail(c,false)}
func(h Orders) AdminDetail(c *gin.Context){h.detail(c,true)}
func(h Orders) transition(c *gin.Context,next string,admin bool){
 err:=h.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB)error{
 var o models.Order
 q:=tx.Clauses(clause.Locking{Strength:"UPDATE"}).Where("id = ?",c.Param("id"))
 if !admin{q=q.Where("user_id = ?",shoppingUser(c))}
 if e:=q.First(&o).Error;e!=nil{return e}
 if o.Status==next{return nil}
 var p models.Payment
 if e:=tx.Clauses(clause.Locking{Strength:"UPDATE"}).Where("order_id = ?",o.ID).First(&p).Error;e!=nil{return e}
 if !allowedOrderTransition(o.Status,next,p.Method,p.Status,admin){return errOrderState}
 switch next{
 case "cancelled":
 if (o.Status!="pending" && o.Status!="confirmed") || p.Status!="pending"{return errOrderState}
 var items []models.OrderItem
 if e:=tx.Where("order_id = ?",o.ID).Order("product_id ASC").Find(&items).Error;e!=nil{return e}
 for _,i:=range items{if _,e:=services.AdjustTx(tx,i.ProductID,i.Quantity,"order cancellation");e!=nil{return e}}
 if e:=tx.Model(&p).Update("status","cancelled").Error;e!=nil{return e}
 case "confirmed":
 if !admin || o.Status!="pending" || p.Method!="cod" || p.Status!="pending"{return errOrderState}
 case "packed":
 if !admin || o.Status!="confirmed"{return errOrderState}
 case "shipped":
 if !admin || o.Status!="packed"{return errOrderState}
 case "delivered":
 if !admin || o.Status!="shipped"{return errOrderState}
 if p.Method=="cod" && p.Status=="pending"{if e:=tx.Model(&p).Update("status","paid").Error;e!=nil{return e}}else if p.Status!="paid"{return errOrderState}
 default:return errOrderState
 }
 previous:=o.Status
 if e:=tx.Model(&o).Update("status",next).Error;e!=nil{return e}
 if e:=tx.Create(&models.OrderEvent{OrderID:o.ID,FromStatus:previous,ToStatus:next,ActorID:shoppingUser(c)}).Error;e!=nil{return e}
 return tx.Create(&models.AuditLog{UserID:shoppingUser(c),Action:"order_"+next,Resource:fmt.Sprintf("order:%d",o.ID)}).Error
 })
 if errors.Is(err,errOrderState){c.JSON(409,gin.H{"error":"order transition not allowed; paid or submitted payments require review"});return}
 if err!=nil{catalogError(c,err);return};c.JSON(200,gin.H{"status":next})
}
func(h Orders) Cancel(c *gin.Context){h.transition(c,"cancelled",false)}
func(h Orders) AdminStatus(c *gin.Context){
 var in struct{Status string `json:"status" binding:"required,oneof=confirmed packed shipped delivered cancelled"`}
 if c.ShouldBindJSON(&in)!=nil{c.JSON(400,gin.H{"error":"invalid status"});return}
 h.transition(c,in.Status,true)
}

func allowedOrderTransition(current,next,method,payment string,admin bool)bool{
 switch next{
 case "cancelled":return (current=="pending" || current=="confirmed") && payment=="pending"
 case "confirmed":return admin && current=="pending" && method=="cod" && payment=="pending"
 case "packed":return admin && current=="confirmed" && (payment=="paid" || (method=="cod" && payment=="pending"))
 case "shipped":return admin && current=="packed" && (payment=="paid" || (method=="cod" && payment=="pending"))
 case "delivered":return admin && current=="shipped" && (payment=="paid" || (method=="cod" && payment=="pending"))
 }
 return false
}
