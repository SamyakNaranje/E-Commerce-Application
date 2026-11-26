package controllers
import (
 "strconv"
 "strings"
 "github.com/gin-gonic/gin"
 "gorm.io/gorm"
 "gorm.io/gorm/clause"
 "github.com/SamyakNaranje/E-Commerce-Application/models"
)
type Reviews struct{DB *gorm.DB}
func(h Reviews) List(c *gin.Context){
 id,e:=strconv.ParseUint(c.Param("id"),10,64)
 page,pe:=strconv.Atoi(c.DefaultQuery("page","1"))
 if e!=nil||id==0||pe!=nil||page<1||page>100000{c.JSON(400,gin.H{"error":"invalid request"});return}
 var product models.Product
 if e:=h.DB.Where("id = ? AND active = ?",id,true).First(&product).Error;e!=nil{catalogError(c,e);return}
 var stats struct{Average float64; Count int64}
 if e:=h.DB.Model(&models.Review{}).Where("product_id = ? AND hidden = ?",id,false).Select("COALESCE(AVG(rating),0) AS average, COUNT(*) AS count").Scan(&stats).Error;e!=nil{catalogError(c,e);return}
 rows:=[]models.Review{}
 if e:=h.DB.Where("product_id = ? AND hidden = ?",id,false).Order("id DESC").Limit(20).Offset((page-1)*20).Find(&rows).Error;e!=nil{catalogError(c,e);return}
 items:=[]gin.H{}
 for _,r:=range rows{items=append(items,gin.H{"id":r.ID,"rating":r.Rating,"comment":r.Comment,"created_at":r.CreatedAt,"verified_purchase":true})}
 c.JSON(200,gin.H{"items":items,"average_rating":stats.Average,"total":stats.Count,"page":page})
}
func(h Reviews) Save(c *gin.Context){
 id,e:=strconv.ParseUint(c.Param("id"),10,64)
 var in struct{Rating int `json:"rating" binding:"required,gte=1,lte=5"`;Comment string `json:"comment" binding:"max=2000"`}
 if e!=nil||id==0||c.ShouldBindJSON(&in)!=nil{c.JSON(400,gin.H{"error":"invalid review"});return}
 uid:=shoppingUser(c)
 var count int64
 if e:=h.DB.Table("order_items").Joins("JOIN orders ON orders.id = order_items.order_id").Where("orders.user_id = ? AND orders.status = ? AND order_items.product_id = ?",uid,"delivered",id).Count(&count).Error;e!=nil{catalogError(c,e);return}
 if count==0{c.JSON(403,gin.H{"error":"verified purchase required"});return}
 row:=models.Review{UserID:uid,ProductID:id,Rating:in.Rating,Comment:strings.TrimSpace(in.Comment)}
 if e:=h.DB.Clauses(clause.OnConflict{Columns:[]clause.Column{{Name:"user_id"},{Name:"product_id"}},DoUpdates:clause.AssignmentColumns([]string{"rating","comment","updated_at"})}).Create(&row).Error;e!=nil{catalogError(c,e);return}
 c.Status(204)
}
func(h Reviews) Delete(c *gin.Context){
 if e:=h.DB.Where("product_id = ? AND user_id = ?",c.Param("id"),shoppingUser(c)).Delete(&models.Review{}).Error;e!=nil{catalogError(c,e);return};c.Status(204)
}
func(h Reviews) AdminList(c *gin.Context){
 rows:=[]models.Review{}
 page,e:=strconv.Atoi(c.DefaultQuery("page","1"))
 if e!=nil||page<1||page>100000{c.JSON(400,gin.H{"error":"invalid page"});return}
 if e:=h.DB.Order("id DESC").Limit(50).Offset((page-1)*50).Find(&rows).Error;e!=nil{catalogError(c,e);return}
 items:=[]gin.H{}
 for _,r:=range rows{items=append(items,gin.H{"review":r,"hidden":r.Hidden})}
 c.JSON(200,items)
}
func(h Reviews) Moderate(c *gin.Context){
 var in struct{Hidden *bool `json:"hidden" binding:"required"`}
 if c.ShouldBindJSON(&in)!=nil{c.JSON(400,gin.H{"error":"invalid moderation"});return}
 err:=h.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB)error{
 r:=tx.Model(&models.Review{}).Where("id = ?",c.Param("id")).Update("hidden",*in.Hidden)
 if r.Error!=nil{return r.Error};if r.RowsAffected==0{return gorm.ErrRecordNotFound}
 return tx.Create(&models.AuditLog{UserID:shoppingUser(c),Action:"review_moderation",Resource:c.Param("id")}).Error
 })
 if err!=nil{catalogError(c,err);return};c.Status(204)
}
