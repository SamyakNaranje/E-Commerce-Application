package routes

import (
 "net/http"
 "github.com/SamyakNaranje/E-Commerce-Application/controllers"
 "github.com/SamyakNaranje/E-Commerce-Application/services"
 "github.com/SamyakNaranje/E-Commerce-Application/middleware"
 "time"
 "context"
 "github.com/gin-gonic/gin"
 "gorm.io/gorm"
)

func Register(r *gin.Engine, db *gorm.DB) {
 a:=services.Auth{DB:db}
 h:=controllers.Auth{Service:a}
 r.POST("/api/auth/register",h.Register)
 r.POST("/api/auth/login",h.Login)
 r.POST("/api/auth/logout",h.Logout)
 r.GET("/api/auth/me",middleware.Authenticate(a),h.Me)
 r.GET("/api/admin/me",middleware.Authenticate(a),middleware.Admin,h.Me)
 r.GET("/",func(c *gin.Context){c.File("views/storefront.html")})
 r.GET("/health",func(c *gin.Context){
  ctx,cancel:=context.WithTimeout(c.Request.Context(), 2*time.Second)
  defer cancel()
  sqlDB,err:=db.DB()
  if err!=nil || sqlDB.PingContext(ctx)!=nil { c.JSON(http.StatusServiceUnavailable,gin.H{"status":"unavailable"});return }
  c.JSON(http.StatusOK,gin.H{"status":"ok"})
 })
}
