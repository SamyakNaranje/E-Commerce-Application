package controllers

import (
 "errors"
 "net/url"
 "strconv"
 "strings"
 "github.com/gin-gonic/gin"
 "gorm.io/gorm"
 "github.com/SamyakNaranje/E-Commerce-Application/models"
)
type Catalog struct { DB *gorm.DB }
type productInput struct {
 Name string `json:"name" binding:"required,max=200"`
 SKU string `json:"sku" binding:"required,max=80"`
 Description string `json:"description" binding:"max=10000"`
 CategoryID uint64 `json:"category_id" binding:"required"`
 PricePaise int64 `json:"price_paise" binding:"gte=0,lte=1000000000"`
}
func (h Catalog) List(c *gin.Context){
 q:=h.DB.WithContext(c.Request.Context()).Model(&models.Product{}).Where("active = ?",true)
 if s:=strings.TrimSpace(c.Query("search"));s!="" {q=q.Where("name ILIKE ? OR description ILIKE ?","%"+s+"%","%"+s+"%")}
 if s:=c.Query("category_id");s!="" {id,e:=strconv.ParseUint(s,10,64);if e!=nil {c.JSON(400,gin.H{"error":"invalid category"});return};q=q.Where("category_id = ?",id)}
 sort:=map[string]string{"price_asc":"price_paise ASC, id ASC","price_desc":"price_paise DESC, id ASC","name":"name ASC, id ASC"}[c.Query("sort")]
 if sort=="" {sort="id DESC"}
 page,e:=strconv.Atoi(c.DefaultQuery("page","1"));if e!=nil || page<1 || page>100000 {c.JSON(400,gin.H{"error":"invalid page"});return}
 var total int64
 if q.Count(&total).Error!=nil {c.JSON(500,gin.H{"error":"catalog unavailable"});return}
 items:=[]models.Product{}
 if q.Preload("Category").Order(sort).Limit(20).Offset((page-1)*20).Find(&items).Error!=nil {c.JSON(500,gin.H{"error":"catalog unavailable"});return}
 c.JSON(200,gin.H{"items":items,"total":total,"page":page,"page_size":20})
}
func (h Catalog) Detail(c *gin.Context){
 var p models.Product
 if err:=h.DB.Preload("Category").Where("id = ? AND active = ?",c.Param("id"),true).First(&p).Error;err!=nil {catalogError(c,err);return}
 images:=[]models.ProductImage{}
 var stock models.Inventory
 if h.DB.Where("product_id = ?",p.ID).Find(&images).Error!=nil {c.JSON(500,gin.H{"error":"catalog unavailable"});return}
 err:=h.DB.First(&stock,"product_id = ?",p.ID).Error
 if err!=nil && !errors.Is(err,gorm.ErrRecordNotFound){catalogError(c,err);return}
 c.JSON(200,gin.H{"product":p,"images":images,"available":stock.Quantity>0})
}
func catalogError(c *gin.Context,err error){
 if errors.Is(err,gorm.ErrRecordNotFound){c.JSON(404,gin.H{"error":"not found"});return}
 c.JSON(500,gin.H{"error":"database operation failed"})
}
func (h Catalog) Create(c *gin.Context){
 var in productInput
 if c.ShouldBindJSON(&in)!=nil || strings.TrimSpace(in.Name)=="" || strings.TrimSpace(in.SKU)=="" {c.JSON(400,gin.H{"error":"invalid product"});return}
 p:=models.Product{Name:strings.TrimSpace(in.Name),SKU:strings.TrimSpace(in.SKU),Description:in.Description,CategoryID:in.CategoryID,PricePaise:in.PricePaise,Active:true}
 err:=h.DB.Transaction(func(tx *gorm.DB)error{
 var cat models.Category
 if e:=tx.First(&cat,in.CategoryID).Error;e!=nil{return e}
 if e:=tx.Create(&p).Error;e!=nil{return e}
 return tx.Create(&models.Inventory{ProductID:p.ID,Quantity:0}).Error
 })
 if err!=nil {catalogError(c,err);return};c.JSON(201,p)
}
func (h Catalog) Update(c *gin.Context){
 var in productInput
 if c.ShouldBindJSON(&in)!=nil || strings.TrimSpace(in.Name)=="" || strings.TrimSpace(in.SKU)=="" {c.JSON(400,gin.H{"error":"invalid product"});return}
 var cat models.Category
 if err:=h.DB.First(&cat,in.CategoryID).Error;err!=nil {catalogError(c,err);return}
 r:=h.DB.Model(&models.Product{}).Where("id = ?",c.Param("id")).Updates(map[string]interface{}{"name":strings.TrimSpace(in.Name),"sku":strings.TrimSpace(in.SKU),"description":in.Description,"category_id":in.CategoryID,"price_paise":in.PricePaise})
 if r.Error!=nil {catalogError(c,r.Error);return};if r.RowsAffected==0 {catalogError(c,gorm.ErrRecordNotFound);return};c.JSON(200,gin.H{"message":"updated"})
}
func (h Catalog) Disable(c *gin.Context){
 r:=h.DB.Model(&models.Product{}).Where("id = ?",c.Param("id")).Update("active",false)
 if r.Error!=nil {catalogError(c,r.Error);return};if r.RowsAffected==0 {catalogError(c,gorm.ErrRecordNotFound);return};c.Status(204)
}
func (h Catalog) Categories(c *gin.Context){
 items:=[]models.Category{}
 if err:=h.DB.Order("name ASC").Find(&items).Error;err!=nil {catalogError(c,err);return};c.JSON(200,items)
}
func (h Catalog) CreateCategory(c *gin.Context){
 var in struct {Name string `json:"name" binding:"required,max=120"`}
 if c.ShouldBindJSON(&in)!=nil || strings.TrimSpace(in.Name)=="" {c.JSON(400,gin.H{"error":"invalid category"});return}
 cat:=models.Category{Name:strings.TrimSpace(in.Name)}
 if err:=h.DB.Create(&cat).Error;err!=nil {catalogError(c,err);return};c.JSON(201,cat)
}
func (h Catalog) UpdateCategory(c *gin.Context){
 var in struct {Name string `json:"name" binding:"required,max=120"`}
 if c.ShouldBindJSON(&in)!=nil || strings.TrimSpace(in.Name)=="" {c.JSON(400,gin.H{"error":"invalid category"});return}
 r:=h.DB.Model(&models.Category{}).Where("id = ?",c.Param("id")).Update("name",strings.TrimSpace(in.Name))
 if r.Error!=nil {catalogError(c,r.Error);return};if r.RowsAffected==0 {catalogError(c,gorm.ErrRecordNotFound);return};c.Status(204)
}

func(h Catalog) AddImage(c *gin.Context){
 var in struct{URL string `json:"url" binding:"required,max=2048"`}
 if c.ShouldBindJSON(&in)!=nil{c.JSON(400,gin.H{"error":"invalid image"});return}
 u,e:=url.Parse(in.URL)
 if e!=nil||u.Scheme!="https"||u.Hostname()==""||u.User!=nil{c.JSON(400,gin.H{"error":"HTTPS image URL required"});return}
 var p models.Product
 if e:=h.DB.First(&p,"id = ?",c.Param("id")).Error;e!=nil{catalogError(c,e);return}
 img:=models.ProductImage{ProductID:p.ID,URL:u.String()}
 if e:=h.DB.Create(&img).Error;e!=nil{catalogError(c,e);return};c.JSON(201,img)
}
func(h Catalog) DeleteImage(c *gin.Context){
 r:=h.DB.Where("id = ? AND product_id = ?",c.Param("image"),c.Param("id")).Delete(&models.ProductImage{})
 if r.Error!=nil{catalogError(c,r.Error);return};if r.RowsAffected==0{catalogError(c,gorm.ErrRecordNotFound);return};c.Status(204)
}
