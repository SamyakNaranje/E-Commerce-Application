package main

import (
 "log"
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
 r:=gin.Default()
 routes.Register(r,db)
 port:=os.Getenv("PORT")
 if port=="" { port="8080" }
 log.Fatal(r.Run(":"+port))
}
