package main

import (
 "log"
 "net/http"
 "time"
 "os"
 "strings"
 "net/url"
 "github.com/gin-gonic/gin"
 "github.com/SamyakNaranje/E-Commerce-Application/config"
 "github.com/SamyakNaranje/E-Commerce-Application/routes"
)

func main() {
 db,err:=config.Database()
 if err!=nil { log.Fatal(err) }
 sqlDB,err:=db.DB()
 if err!=nil { log.Fatal(err) }
 defer sqlDB.Close()
 if len(os.Getenv("JWT_SECRET"))<32 {log.Fatal("JWT_SECRET must contain at least 32 bytes")}
 production:=os.Getenv("APP_ENV")=="production"
 origin:=os.Getenv("APP_ORIGIN")
 if production{u,e:=url.Parse(origin);if e!=nil||u.Scheme!="https"||u.Host==""||u.User!=nil||u.Path!=""||u.RawQuery!=""||u.Fragment!=""||os.Getenv("COOKIE_SECURE")=="false"{log.Fatal("production requires HTTPS APP_ORIGIN and secure cookies")};gin.SetMode(gin.ReleaseMode)}
 r:=gin.New();r.Use(gin.CustomRecovery(func(c *gin.Context,_ interface{}){log.Print("request panic recovered");c.AbortWithStatusJSON(500,gin.H{"error":"internal server error"})}))
 if err:=r.SetTrustedProxies(nil);err!=nil{log.Fatal(err)}
 if production{
 proxies:=os.Getenv("TRUSTED_PROXIES");if proxies!=""{if err:=r.SetTrustedProxies(strings.Split(proxies,","));err!=nil{log.Fatal("invalid trusted proxies")}}
 }
 routes.Register(r,db)
 port:=os.Getenv("PORT")
 if port=="" { port="8080" }
 bind:=os.Getenv("BIND_ADDRESS");if bind==""{bind="127.0.0.1"}
 server:=&http.Server{Addr:bind+":"+port,Handler:r,ReadHeaderTimeout:5*time.Second,ReadTimeout:15*time.Second,WriteTimeout:30*time.Second,IdleTimeout:60*time.Second,MaxHeaderBytes:32<<10}
 log.Fatal(server.ListenAndServe())
}
