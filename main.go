package main

import (
 "log"
 "net/http"
 "time"
 "os"
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
 r:=gin.Default()
 if err:=r.SetTrustedProxies(nil);err!=nil{log.Fatal(err)}
 routes.Register(r,db)
 port:=os.Getenv("PORT")
 if port=="" { port="8080" }
 server:=&http.Server{Addr:":"+port,Handler:r,ReadHeaderTimeout:5*time.Second,ReadTimeout:15*time.Second,WriteTimeout:30*time.Second,IdleTimeout:60*time.Second,MaxHeaderBytes:1<<20}
 log.Fatal(server.ListenAndServe())
}
