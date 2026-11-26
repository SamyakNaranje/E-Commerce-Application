package routes

import (
 "net/http"
 "time"
 "context"
 "github.com/gin-gonic/gin"
 "gorm.io/gorm"
)

func Register(r *gin.Engine, db *gorm.DB) {
 r.GET("/",func(c *gin.Context){c.File("views/storefront.html")})
 r.GET("/health",func(c *gin.Context){
  ctx,cancel:=context.WithTimeout(c.Request.Context(), 2*time.Second)
  defer cancel()
  sqlDB,err:=db.DB()
  if err!=nil || sqlDB.PingContext(ctx)!=nil { c.JSON(http.StatusServiceUnavailable,gin.H{"status":"unavailable"});return }
  c.JSON(http.StatusOK,gin.H{"status":"ok"})
 })
}
