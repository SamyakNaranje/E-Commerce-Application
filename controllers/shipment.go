package controllers

import (
 "errors"
 "time"
 "strings"
 "github.com/gin-gonic/gin"
 "gorm.io/gorm"
 "gorm.io/gorm/clause"
 "github.com/SamyakNaranje/E-Commerce-Application/models"
)
func(h Delivery) Shipment(c *gin.Context){
 var o models.Order
 if e:=h.DB.Where("id = ? AND user_id = ?",c.Param("id"),shoppingUser(c)).First(&o).Error;e!=nil{catalogError(c,e);return}
 var s models.Shipment
 e:=h.DB.Where("order_id = ?",o.ID).First(&s).Error
 if errors.Is(e,gorm.ErrRecordNotFound){c.JSON(200,gin.H{"status":o.Status,"shipment":nil});return}
 if e!=nil{catalogError(c,e);return}
 c.JSON(200,gin.H{"status":o.Status,"shipment":s})
}
func(h Delivery) SaveShipment(c *gin.Context){
 var in struct{
 Courier string `json:"courier" binding:"max=120"`
 TrackingNumber string `json:"tracking_number" binding:"max=120"`
 ExpectedDelivery *time.Time `json:"expected_delivery"`
 }
 if c.ShouldBindJSON(&in)!=nil || (in.ExpectedDelivery!=nil && in.ExpectedDelivery.Before(time.Now().Add(-24*time.Hour))){c.JSON(400,gin.H{"error":"invalid shipment"});return}
 err:=h.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB)error{
 var o models.Order
 if e:=tx.Clauses(clause.Locking{Strength:"UPDATE"}).First(&o,"id = ?",c.Param("id")).Error;e!=nil{return e}
 if o.Status!="confirmed" && o.Status!="packed" && o.Status!="shipped"{return errOrderState}
 s:=models.Shipment{OrderID:o.ID,Courier:strings.TrimSpace(in.Courier),TrackingNumber:strings.TrimSpace(in.TrackingNumber),ExpectedDelivery:in.ExpectedDelivery}
 if e:=tx.Clauses(clause.OnConflict{Columns:[]clause.Column{{Name:"order_id"}},DoUpdates:clause.AssignmentColumns([]string{"courier","tracking_number","expected_delivery","updated_at"})}).Create(&s).Error;e!=nil{return e}
 return tx.Create(&models.AuditLog{UserID:shoppingUser(c),Action:"shipment_update",Resource:c.Param("id")}).Error
 })
 if errors.Is(err,errOrderState){c.JSON(409,gin.H{"error":"shipment cannot be updated in this order state"});return}
 if err!=nil{catalogError(c,err);return};c.Status(204)
}
