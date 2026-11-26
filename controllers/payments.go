package controllers

import (
 "errors"
 "fmt"
 "net/url"
 "os"
 "strings"
 "github.com/gin-gonic/gin"
 "gorm.io/gorm"
 "gorm.io/gorm/clause"
 "github.com/SamyakNaranje/E-Commerce-Application/models"
)
type Payments struct{DB *gorm.DB}
var errPaymentState=errors.New("invalid payment state")
func (h Payments) Details(c *gin.Context){
 var order models.Order
 if err:=h.DB.Where("id = ? AND user_id = ?",c.Param("id"),shoppingUser(c)).First(&order).Error;err!=nil{catalogError(c,err);return}
 var p models.Payment
 if err:=h.DB.Where("order_id = ?",order.ID).First(&p).Error;err!=nil{catalogError(c,err);return}
 out:=gin.H{"payment":p,"order_status":order.Status}
 if p.Method=="upi" && p.Status=="pending" && order.Status=="pending" && p.AmountPaise>0 {
 vpa:=strings.TrimSpace(os.Getenv("UPI_ID"))
 if vpa==""{c.JSON(503,gin.H{"error":"UPI not configured"});return}
 q:=url.Values{"pa":{vpa},"pn":{os.Getenv("UPI_PAYEE_NAME")},"am":{fmt.Sprintf("%d.%02d",p.AmountPaise/100,p.AmountPaise%100)},"cu":{"INR"},"tn":{fmt.Sprintf("Order %d",order.ID)}}
 out["upi_uri"]="upi://pay?"+q.Encode()
 out["upi_id"]=vpa
 }
 c.JSON(200,out)
}
func (h Payments) Reference(c *gin.Context){
 var in struct{Reference string `json:"reference" binding:"required,min=6,max=100"`}
 if c.ShouldBindJSON(&in)!=nil || len(strings.TrimSpace(in.Reference))<6{c.JSON(400,gin.H{"error":"invalid reference"});return}
 err:=h.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB)error{
 var o models.Order
 if e:=tx.Clauses(clause.Locking{Strength:"UPDATE"}).Where("id = ? AND user_id = ?",c.Param("id"),shoppingUser(c)).First(&o).Error;e!=nil{return e}
 var p models.Payment
 if e:=tx.Clauses(clause.Locking{Strength:"UPDATE"}).Where("order_id = ?",o.ID).First(&p).Error;e!=nil{return e}
 if o.Status!="pending" || p.Method!="upi" || p.Status!="pending"{return errPaymentState}
 return tx.Model(&p).Updates(map[string]interface{}{"reference":strings.TrimSpace(in.Reference),"status":"submitted"}).Error
 })
 if errors.Is(err,errPaymentState){c.JSON(409,gin.H{"error":"payment cannot be submitted"});return}
 if err!=nil{catalogError(c,err);return};c.JSON(200,gin.H{"status":"submitted"})
}
func (h Payments) Verify(c *gin.Context){
 var in struct{Decision string `json:"decision" binding:"required,oneof=approve reject"`;Note string `json:"note" binding:"required,max=200"`}
 if c.ShouldBindJSON(&in)!=nil || strings.TrimSpace(in.Note)==""{c.JSON(400,gin.H{"error":"invalid decision"});return}
 err:=h.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB)error{
 var o models.Order
 if e:=tx.Clauses(clause.Locking{Strength:"UPDATE"}).First(&o,"id = ?",c.Param("id")).Error;e!=nil{return e}
 var p models.Payment
 if e:=tx.Clauses(clause.Locking{Strength:"UPDATE"}).Where("order_id = ?",o.ID).First(&p).Error;e!=nil{return e}
 if p.Method!="upi" || p.Status!="submitted" || o.Status!="pending"{return errPaymentState}
 status:="pending"
 if in.Decision=="approve"{status="paid"}
 if e:=tx.Model(&p).Update("status",status).Error;e!=nil{return e}
 if status=="paid"{if e:=tx.Model(&o).Update("status","confirmed").Error;e!=nil{return e}}
 return tx.Create(&models.AuditLog{UserID:shoppingUser(c),Action:"payment_"+in.Decision,Resource:fmt.Sprintf("order:%d %s",o.ID,in.Note)}).Error
 })
 if errors.Is(err,errPaymentState){c.JSON(409,gin.H{"error":"invalid payment state"});return}
 if err!=nil{catalogError(c,err);return};c.Status(204)
}
func (h Payments) List(c *gin.Context){
 rows:=[]models.Payment{}
 q:=h.DB.Order("id DESC").Limit(200)
 if s:=c.Query("status");s!=""{q=q.Where("status = ?",s)}
 if err:=q.Find(&rows).Error;err!=nil{catalogError(c,err);return};c.JSON(200,rows)
}
