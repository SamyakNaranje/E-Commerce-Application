package controllers

import (
 "strings"
 "regexp"
 "github.com/gin-gonic/gin"
 "gorm.io/gorm"
 "github.com/SamyakNaranje/E-Commerce-Application/models"
)
type Delivery struct{DB *gorm.DB}
var pinPattern=regexp.MustCompile("^[1-9][0-9]{5}$")
var phonePattern=regexp.MustCompile("^[6-9][0-9]{9}$")
type addressInput struct{
 Recipient string `json:"recipient" binding:"required,max=120"`
 Phone string `json:"phone" binding:"required"`
 Line string `json:"line" binding:"required,max=500"`
 City string `json:"city" binding:"required,max=120"`
 PostalCode string `json:"postal_code" binding:"required"`
}
func(h Delivery) Addresses(c *gin.Context){
 rows:=[]models.Address{}
 if e:=h.DB.Where("user_id = ?",shoppingUser(c)).Order("id DESC").Find(&rows).Error;e!=nil{catalogError(c,e);return};c.JSON(200,rows)
}
func(h Delivery) SaveAddress(c *gin.Context){
 var in addressInput
 if c.ShouldBindJSON(&in)!=nil || !pinPattern.MatchString(in.PostalCode)||!phonePattern.MatchString(in.Phone)||strings.TrimSpace(in.Recipient)==""||strings.TrimSpace(in.Line)==""||strings.TrimSpace(in.City)==""{c.JSON(400,gin.H{"error":"invalid address"});return}
 a:=models.Address{UserID:shoppingUser(c),Recipient:strings.TrimSpace(in.Recipient),Phone:in.Phone,Line:strings.TrimSpace(in.Line),City:strings.TrimSpace(in.City),PostalCode:in.PostalCode}
 if id:=c.Param("id");id!=""{
 r:=h.DB.Model(&models.Address{}).Where("id = ? AND user_id = ?",id,a.UserID).Updates(map[string]interface{}{"recipient":a.Recipient,"phone":a.Phone,"line":a.Line,"city":a.City,"postal_code":a.PostalCode})
 if r.Error!=nil{catalogError(c,r.Error);return};if r.RowsAffected==0{catalogError(c,gorm.ErrRecordNotFound);return};c.Status(204);return
 }
 if e:=h.DB.Create(&a).Error;e!=nil{catalogError(c,e);return};c.JSON(201,a)
}
func(h Delivery) DeleteAddress(c *gin.Context){
 r:=h.DB.Where("id = ? AND user_id = ?",c.Param("id"),shoppingUser(c)).Delete(&models.Address{})
 if r.Error!=nil{catalogError(c,r.Error);return};if r.RowsAffected==0{catalogError(c,gorm.ErrRecordNotFound);return};c.Status(204)
}
func(h Delivery) Check(c *gin.Context){
 pin:=c.Param("pin")
 if !pinPattern.MatchString(pin){c.JSON(400,gin.H{"error":"invalid PIN"});return}
 var a models.DeliveryArea
 r:=h.DB.Where("postal_code = ? AND active = ?",pin,true).Limit(1).Find(&a)
 if r.Error!=nil{catalogError(c,r.Error);return}
 if r.RowsAffected==0{c.JSON(200,gin.H{"serviceable":false});return}
 c.JSON(200,gin.H{"serviceable":true,"charge_paise":a.ChargePaise})
}
func(h Delivery) Areas(c *gin.Context){
 rows:=[]models.DeliveryArea{}
 if e:=h.DB.Order("postal_code").Limit(1000).Find(&rows).Error;e!=nil{catalogError(c,e);return};c.JSON(200,rows)
}
func(h Delivery) SaveArea(c *gin.Context){
 var in struct{PostalCode string `json:"postal_code" binding:"required"`;ChargePaise *int64 `json:"charge_paise" binding:"required"`;Active *bool `json:"active" binding:"required"`}
 if c.ShouldBindJSON(&in)!=nil || !pinPattern.MatchString(in.PostalCode)||*in.ChargePaise<0||*in.ChargePaise>10000000{c.JSON(400,gin.H{"error":"invalid delivery area"});return}
 a:=models.DeliveryArea{PostalCode:in.PostalCode,ChargePaise:*in.ChargePaise,Active:*in.Active}
 if id:=c.Param("id");id!=""{
 r:=h.DB.Model(&models.DeliveryArea{}).Where("id = ?",id).Updates(map[string]interface{}{"postal_code":a.PostalCode,"charge_paise":a.ChargePaise,"active":a.Active})
 if r.Error!=nil{catalogError(c,r.Error);return};if r.RowsAffected==0{catalogError(c,gorm.ErrRecordNotFound);return};c.Status(204);return
 }
 if e:=h.DB.Create(&a).Error;e!=nil{catalogError(c,e);return};c.JSON(201,a)
}
