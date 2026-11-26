package controllers

import (
 "errors"
 "strconv"
 "strings"
 "github.com/gin-gonic/gin"
 "github.com/SamyakNaranje/E-Commerce-Application/models"
 "github.com/SamyakNaranje/E-Commerce-Application/services"
)
type Inventory struct { Service services.Inventory }
func (h Inventory) List(c *gin.Context){
 q:=h.Service.DB.WithContext(c.Request.Context()).Model(&models.Inventory{})
 switch c.Query("status"){
 case "low_stock":q=q.Where("quantity BETWEEN 1 AND 9")
 case "out_of_stock":q=q.Where("quantity = 0")
 case "in_stock":q=q.Where("quantity >= 10")
 case "":
 default:c.JSON(400,gin.H{"error":"invalid status"});return
 }
 page,e:=strconv.Atoi(c.DefaultQuery("page","1"))
 if e!=nil || page<1 || page>100000{c.JSON(400,gin.H{"error":"invalid page"});return}
 var total int64
 if err:=q.Count(&total).Error;err!=nil{catalogError(c,err);return}
 rows:=[]models.Inventory{}
 if err:=q.Order("quantity ASC, product_id ASC").Limit(50).Offset((page-1)*50).Find(&rows).Error;err!=nil{catalogError(c,err);return}
 items:=[]gin.H{}
 for _,v:=range rows{items=append(items,gin.H{"product_id":v.ProductID,"quantity":v.Quantity,"status":services.StockStatus(v.Quantity)})}
 c.JSON(200,gin.H{"items":items,"total":total,"page":page})
}
func (h Inventory) Adjust(c *gin.Context){
 id,e:=strconv.ParseUint(c.Param("id"),10,64)
 var in struct{Delta int64 `json:"delta" binding:"required"`;Reason string `json:"reason" binding:"required,max=200"`}
 if e!=nil || id==0 || c.ShouldBindJSON(&in)!=nil || strings.TrimSpace(in.Reason)==""{c.JSON(400,gin.H{"error":"invalid adjustment"});return}
 u:=c.MustGet("user").(models.User)
 stock,err:=h.Service.Adjust(c.Request.Context(),id,in.Delta,strings.TrimSpace(in.Reason),u.ID)
 if errors.Is(err,services.ErrInvalidStock){c.JSON(409,gin.H{"error":"stock adjustment exceeds allowed quantity"});return}
 if err!=nil{catalogError(c,err);return}
 c.JSON(200,gin.H{"product_id":id,"quantity":stock.Quantity,"status":services.StockStatus(stock.Quantity)})
}
func (h Inventory) History(c *gin.Context){
 id,e:=strconv.ParseUint(c.Param("id"),10,64)
 page,pErr:=strconv.Atoi(c.DefaultQuery("page","1"))
 if e!=nil || id==0 || pErr!=nil || page<1 || page>100000{c.JSON(400,gin.H{"error":"invalid request"});return}
 rows:=[]models.StockMovement{}
 if err:=h.Service.DB.WithContext(c.Request.Context()).Where("product_id = ?",id).Order("id DESC").Limit(50).Offset((page-1)*50).Find(&rows).Error;err!=nil{catalogError(c,err);return}
 c.JSON(200,rows)
}
