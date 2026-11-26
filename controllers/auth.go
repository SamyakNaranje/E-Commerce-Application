package controllers

import (
 "net/http"
 "os"
 "github.com/gin-gonic/gin"
 "github.com/SamyakNaranje/E-Commerce-Application/services"
)
type Auth struct { Service services.Auth }
type credentials struct {
 Name string `json:"name"`
 Email string `json:"email" binding:"required"`
 Password string `json:"password" binding:"required"`
}
func (a Auth) Register(c *gin.Context){
 var in credentials
 if c.ShouldBindJSON(&in)!=nil || a.Service.Register(in.Name,in.Email,in.Password)!=nil {c.JSON(400,gin.H{"error":"registration failed"});return}
 c.JSON(201,gin.H{"message":"registered"})
}
func cookie(c *gin.Context,value string,age int){
 c.SetSameSite(http.SameSiteStrictMode)
 c.SetCookie("session",value,age,"/","",os.Getenv("COOKIE_SECURE")!="false",true)
}
func (a Auth) Login(c *gin.Context){
 var in credentials
 if c.ShouldBindJSON(&in)!=nil {c.JSON(400,gin.H{"error":"invalid request"});return}
 token,err:=a.Service.Login(in.Email,in.Password)
 if err!=nil {c.JSON(401,gin.H{"error":"login failed"});return}
 cookie(c,token,1800);c.JSON(200,gin.H{"message":"logged in"})
}
func (a Auth) Logout(c *gin.Context){cookie(c,"",-1);c.JSON(200,gin.H{"message":"logged out"})}
func (a Auth) Me(c *gin.Context){v,_:=c.Get("user");c.JSON(200,v)}
