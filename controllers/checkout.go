package controllers

import (
 "errors"
 "fmt"
 "strings"
 "time"
 "github.com/gin-gonic/gin"
 "gorm.io/gorm"
 "gorm.io/gorm/clause"
 "github.com/SamyakNaranje/E-Commerce-Application/models"
 "github.com/SamyakNaranje/E-Commerce-Application/services"
)
type Checkout struct{DB *gorm.DB}
var errCheckout=errors.New("cart, address, coupon or stock invalid")
func (h Checkout) Place(c *gin.Context){
 var in struct{
 AddressID uint64 `json:"address_id" binding:"required"`
 Coupon string `json:"coupon" binding:"max=80"`
 Method string `json:"method" binding:"required,oneof=upi cod"`
 Key string `json:"idempotency_key" binding:"required,min=16,max=100"`
 }
 if c.ShouldBindJSON(&in)!=nil{c.JSON(400,gin.H{"error":"invalid checkout"});return}
 uid:=shoppingUser(c)
 var order models.Order
 err:=h.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB)error{
 var u models.User
 if e:=tx.Clauses(clause.Locking{Strength:"UPDATE"}).First(&u,uid).Error;e!=nil{return e}
 e:=tx.Where("user_id = ? AND checkout_key = ?",uid,in.Key).First(&order).Error
 if e==nil{return nil};if !errors.Is(e,gorm.ErrRecordNotFound){return e}
 var addr models.Address
 if e:=tx.Where("id = ? AND user_id = ?",in.AddressID,uid).First(&addr).Error;e!=nil{return e}
 var area models.DeliveryArea
 if e:=tx.Where("postal_code = ? AND active = ?",addr.PostalCode,true).First(&area).Error;e!=nil{return e}
 var cart []models.CartItem
 if e:=tx.Where("user_id = ?",uid).Order("product_id ASC").Find(&cart).Error;e!=nil{return e}
 if len(cart)==0 || len(cart)>1000{return errCheckout}
 var subtotal int64
 items:=[]models.OrderItem{}
 for _,row:=range cart{
 var p models.Product
 if e:=tx.Clauses(clause.Locking{Strength:"UPDATE"}).Where("id = ? AND active = ?",row.ProductID,true).First(&p).Error;e!=nil{return e}
 if row.Quantity<1 || row.Quantity>1000 || p.PricePaise<0 || p.PricePaise>1000000000{return errCheckout}
 if _,e:=services.AdjustTx(tx,p.ID,-row.Quantity,"checkout");e!=nil{return e}
 subtotal+=row.Quantity*p.PricePaise
 items=append(items,models.OrderItem{ProductID:p.ID,ProductName:p.Name,Quantity:row.Quantity,UnitPricePaise:p.PricePaise})
 }
 var discount int64
 code:=strings.ToUpper(strings.TrimSpace(in.Coupon))
 if code!=""{
 var coupon models.Coupon
 if e:=tx.Where("code = ? AND active = ? AND expires_at > ?",code,true,time.Now()).First(&coupon).Error;e!=nil{return e}
 if coupon.DiscountPercent<1 || coupon.DiscountPercent>100{return errCheckout}
 discount=subtotal*coupon.DiscountPercent/100
 }
 order=models.Order{UserID:uid,CheckoutKey:in.Key,Status:"pending",TotalPaise:subtotal-discount+area.ChargePaise,SubtotalPaise:subtotal,DiscountPaise:discount,DeliveryChargePaise:area.ChargePaise,CouponCode:code,DeliveryAddress:fmt.Sprintf("%s, %s, %s, %s, %s",addr.Recipient,addr.Phone,addr.Line,addr.City,addr.PostalCode)}
 if e:=tx.Create(&order).Error;e!=nil{return e}
 for i:=range items{items[i].OrderID=order.ID}
 if e:=tx.Create(&items).Error;e!=nil{return e}
 if e:=tx.Create(&models.Payment{OrderID:order.ID,Method:in.Method,Status:"pending",AmountPaise:order.TotalPaise}).Error;e!=nil{return e}
 return tx.Where("user_id = ?",uid).Delete(&models.CartItem{}).Error
 })
 if errors.Is(err,errCheckout)||errors.Is(err,services.ErrInvalidStock)||errors.Is(err,gorm.ErrRecordNotFound){c.JSON(409,gin.H{"error":"checkout unavailable: verify cart, stock, address and coupon"});return}
 if err!=nil{catalogError(c,err);return}
 c.JSON(200,order)
}
func (h Checkout) CreateCoupon(c *gin.Context){
 var in struct{Code string `json:"code" binding:"required,max=80"`;Percent int64 `json:"discount_percent" binding:"required,gte=1,lte=100"`;Expires time.Time `json:"expires_at" binding:"required"`}
 if c.ShouldBindJSON(&in)!=nil || !in.Expires.After(time.Now()) || strings.TrimSpace(in.Code)==""{c.JSON(400,gin.H{"error":"invalid coupon"});return}
 coupon:=models.Coupon{Code:strings.ToUpper(strings.TrimSpace(in.Code)),DiscountPercent:in.Percent,ExpiresAt:in.Expires,Active:true}
 if err:=h.DB.Create(&coupon).Error;err!=nil{catalogError(c,err);return};c.JSON(201,coupon)
}
func (h Checkout) Coupons(c *gin.Context){
 rows:=[]models.Coupon{}
 if err:=h.DB.Order("id DESC").Limit(200).Find(&rows).Error;err!=nil{catalogError(c,err);return};c.JSON(200,rows)
}
func (h Checkout) DisableCoupon(c *gin.Context){
 r:=h.DB.Model(&models.Coupon{}).Where("id = ?",c.Param("id")).Update("active",false)
 if r.Error!=nil{catalogError(c,r.Error);return};if r.RowsAffected==0{catalogError(c,gorm.ErrRecordNotFound);return};c.Status(204)
}
