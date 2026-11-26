package controllers

import (
 "strconv"
 "errors"
 "github.com/gin-gonic/gin"
 "gorm.io/gorm"
 "gorm.io/gorm/clause"
 "github.com/SamyakNaranje/E-Commerce-Application/models"
)
type Shopping struct { DB *gorm.DB }
func shoppingUser(c *gin.Context) uint64 {return c.MustGet("user").(models.User).ID}
func (h Shopping) Cart(c *gin.Context){
 rows:=[]models.CartItem{}
 if err:=h.DB.WithContext(c.Request.Context()).Preload("Product").Where("user_id = ?",shoppingUser(c)).Order("id").Find(&rows).Error;err!=nil{catalogError(c,err);return}
 var total int64
 items:=[]gin.H{}
 for _,r:=range rows{
 var stock models.Inventory
 err:=h.DB.First(&stock,"product_id = ?",r.ProductID).Error
 if err!=nil && !errors.Is(err,gorm.ErrRecordNotFound){catalogError(c,err);return}
 subtotal:=r.Quantity*r.Product.PricePaise
 available:=r.Product.Active && stock.Quantity>=r.Quantity
 if available{total+=subtotal}
 items=append(items,gin.H{"product":r.Product,"quantity":r.Quantity,"subtotal_paise":subtotal,"available":available})
 }
 c.JSON(200,gin.H{"items":items,"available_total_paise":total})
}
func (h Shopping) SetCart(c *gin.Context){
 id,e:=strconv.ParseUint(c.Param("id"),10,64)
 var in struct{Quantity int64 `json:"quantity" binding:"required,gte=1,lte=1000"`}
 if e!=nil || id==0 || c.ShouldBindJSON(&in)!=nil{c.JSON(400,gin.H{"error":"invalid product or quantity"});return}
 uid:=shoppingUser(c)
 err:=h.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB)error{
 var u models.User
 if e:=tx.Clauses(clause.Locking{Strength:"UPDATE"}).First(&u,uid).Error;e!=nil{return e}
 var p models.Product
 if e:=tx.Where("id = ? AND active = ?",id,true).First(&p).Error;e!=nil{return e}
 var stock models.Inventory
 if e:=tx.First(&stock,"product_id = ?",id).Error;e!=nil{return e}
 if stock.Quantity<in.Quantity{return errCartStock}
 return tx.Clauses(clause.OnConflict{Columns:[]clause.Column{{Name:"user_id"},{Name:"product_id"}},DoUpdates:clause.AssignmentColumns([]string{"quantity","updated_at"})}).Create(&models.CartItem{UserID:uid,ProductID:id,Quantity:in.Quantity}).Error
 })
 if errors.Is(err,errCartStock){c.JSON(409,gin.H{"error":"insufficient stock"});return}
 if err!=nil{catalogError(c,err);return};c.Status(204)
}
var errCartStock=errors.New("insufficient stock")
func (h Shopping) RemoveCart(c *gin.Context){
 id,e:=strconv.ParseUint(c.Param("id"),10,64)
 if e!=nil || id==0{c.JSON(400,gin.H{"error":"invalid product"});return}
 uid:=shoppingUser(c)
 err:=h.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB)error{
 var u models.User
 if e:=tx.Clauses(clause.Locking{Strength:"UPDATE"}).First(&u,uid).Error;e!=nil{return e}
 return tx.Where("user_id = ? AND product_id = ?",uid,id).Delete(&models.CartItem{}).Error
 })
 if err!=nil{catalogError(c,err);return};c.Status(204)
}
func (h Shopping) Wishlist(c *gin.Context){
 rows:=[]models.WishlistItem{}
 if err:=h.DB.WithContext(c.Request.Context()).Preload("Product").Where("user_id = ?",shoppingUser(c)).Order("id DESC").Find(&rows).Error;err!=nil{catalogError(c,err);return}
 c.JSON(200,rows)
}
func (h Shopping) AddWishlist(c *gin.Context){
 id,e:=strconv.ParseUint(c.Param("id"),10,64)
 if e!=nil || id==0{c.JSON(400,gin.H{"error":"invalid product"});return}
 var p models.Product
 if err:=h.DB.Where("id = ? AND active = ?",id,true).First(&p).Error;err!=nil{catalogError(c,err);return}
 if err:=h.DB.Clauses(clause.OnConflict{Columns:[]clause.Column{{Name:"user_id"},{Name:"product_id"}},DoNothing:true}).Create(&models.WishlistItem{UserID:shoppingUser(c),ProductID:id}).Error;err!=nil{catalogError(c,err);return}
 c.Status(204)
}
func (h Shopping) RemoveWishlist(c *gin.Context){
 id,e:=strconv.ParseUint(c.Param("id"),10,64)
 if e!=nil || id==0{c.JSON(400,gin.H{"error":"invalid product"});return}
 if err:=h.DB.Where("user_id = ? AND product_id = ?",shoppingUser(c),id).Delete(&models.WishlistItem{}).Error;err!=nil{catalogError(c,err);return};c.Status(204)
}
