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
 catalog:=controllers.Catalog{DB:db}
 customer:=r.Group("/api",middleware.Authenticate(a))
 shopping:=controllers.Shopping{DB:db}
 customer.GET("/cart",shopping.Cart)
 customer.PUT("/cart/:id",shopping.SetCart)
 customer.DELETE("/cart/:id",shopping.RemoveCart)
 customer.GET("/wishlist",shopping.Wishlist)
 customer.PUT("/wishlist/:id",shopping.AddWishlist)
 customer.DELETE("/wishlist/:id",shopping.RemoveWishlist)
 customer.GET("/products",catalog.List)
 customer.GET("/products/:id",catalog.Detail)
 customer.GET("/categories",catalog.Categories)
 admin:=r.Group("/api/admin",middleware.Authenticate(a),middleware.Admin)
 inventory:=controllers.Inventory{Service:services.Inventory{DB:db}}
 admin.GET("/inventory",inventory.List)
 admin.POST("/inventory/:id/adjust",inventory.Adjust)
 admin.GET("/inventory/:id/history",inventory.History)
 checkout:=controllers.Checkout{DB:db}
 payments:=controllers.Payments{DB:db}
 customer.GET("/orders/:id/payment",payments.Details)
 customer.POST("/orders/:id/payment/reference",payments.Reference)
 admin.GET("/payments",payments.List)
 admin.POST("/orders/:id/payment/verify",payments.Verify)
 orders:=controllers.Orders{DB:db}
 customer.GET("/orders",orders.List)
 customer.GET("/orders/:id",orders.Detail)
 customer.POST("/orders/:id/cancel",orders.Cancel)
 admin.GET("/orders",orders.AdminList)
 admin.GET("/orders/:id",orders.AdminDetail)
 admin.PUT("/orders/:id/status",orders.AdminStatus)
 delivery:=controllers.Delivery{DB:db}
 customer.GET("/orders/:id/shipment",delivery.Shipment)
 admin.PUT("/orders/:id/shipment",delivery.SaveShipment)
 customer.GET("/addresses",delivery.Addresses)
 customer.POST("/addresses",delivery.SaveAddress)
 customer.PUT("/addresses/:id",delivery.SaveAddress)
 customer.DELETE("/addresses/:id",delivery.DeleteAddress)
 customer.GET("/delivery/:pin",delivery.Check)
 admin.GET("/delivery-areas",delivery.Areas)
 admin.POST("/delivery-areas",delivery.SaveArea)
 admin.PUT("/delivery-areas/:id",delivery.SaveArea)
 returns:=controllers.Returns{DB:db}
 customer.POST("/returns",returns.Request)
 customer.GET("/returns",returns.List)
 customer.GET("/refunds",returns.Refunds)
 admin.GET("/returns",returns.AdminList)
 admin.PUT("/returns/:id/status",returns.Update)
 admin.GET("/refunds",returns.Refunds)
 reviews:=controllers.Reviews{DB:db}
 customer.GET("/products/:id/reviews",reviews.List)
 customer.PUT("/products/:id/review",reviews.Save)
 customer.DELETE("/products/:id/review",reviews.Delete)
 admin.GET("/reviews",reviews.AdminList)
 admin.PUT("/reviews/:id/moderation",reviews.Moderate)
 dashboard:=controllers.Dashboard{DB:db}
 admin.GET("/dashboard",dashboard.Summary)
 admin.GET("/reports/sales",dashboard.Sales)
 admin.GET("/reports/sales.csv",dashboard.CSV)
 admin.GET("/reports/best-sellers",dashboard.BestSellers)
 admin.GET("/customers",dashboard.Customers)
 customer.POST("/checkout",checkout.Place)
 admin.POST("/coupons",checkout.CreateCoupon)
 admin.GET("/coupons",checkout.Coupons)
 admin.DELETE("/coupons/:id",checkout.DisableCoupon)
 admin.POST("/products/:id/images",catalog.AddImage)
 admin.DELETE("/products/:id/images/:image",catalog.DeleteImage)
 admin.POST("/products",catalog.Create)
 admin.PUT("/products/:id",catalog.Update)
 admin.DELETE("/products/:id",catalog.Disable)
 admin.POST("/categories",catalog.CreateCategory)
 admin.PUT("/categories/:id",catalog.UpdateCategory)
 r.StaticFile("/assets/shop.js","assets/shop.js")
 r.StaticFile("/assets/storefront.js","assets/storefront.js")
 r.GET("/",func(c *gin.Context){c.File("views/live-storefront.html")})
 r.GET("/account",func(c *gin.Context){c.File("views/shop.html")})
 r.GET("/legacy",func(c *gin.Context){c.Redirect(http.StatusTemporaryRedirect,"/")})
 r.GET("/health",func(c *gin.Context){
  ctx,cancel:=context.WithTimeout(c.Request.Context(), 2*time.Second)
  defer cancel()
  sqlDB,err:=db.DB()
  if err!=nil || sqlDB.PingContext(ctx)!=nil { c.JSON(http.StatusServiceUnavailable,gin.H{"status":"unavailable"});return }
  c.JSON(http.StatusOK,gin.H{"status":"ok"})
 })
}
