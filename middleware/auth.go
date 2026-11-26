package middleware

import (
 "net/http"
 "github.com/gin-gonic/gin"
 "github.com/SamyakNaranje/E-Commerce-Application/services"
 "github.com/SamyakNaranje/E-Commerce-Application/models"
)
func Authenticate(a services.Auth) gin.HandlerFunc {
 return func(c *gin.Context){
  raw,err:=c.Cookie("session")
  if err!=nil {c.AbortWithStatusJSON(http.StatusUnauthorized,gin.H{"error":"login required"});return}
  u,err:=a.Verify(raw)
  if err!=nil {c.AbortWithStatusJSON(http.StatusUnauthorized,gin.H{"error":"invalid session"});return}
  c.Set("user",u);c.Next()
 }
}
func Admin(c *gin.Context){
 v,ok:=c.Get("user");u,valid:=v.(models.User)
 if !ok || !valid || u.Role!="admin" {c.AbortWithStatusJSON(http.StatusForbidden,gin.H{"error":"admin required"});return}
 c.Next()
}
