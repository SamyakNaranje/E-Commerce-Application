package config

import (
 "os"
 "context"
 "time"
 "gorm.io/gorm/logger"
 "fmt"
 "gorm.io/driver/postgres"
 "gorm.io/gorm"
)

func Database() (*gorm.DB,error) {
 dsn := os.Getenv("DATABASE_URL")
 if dsn == "" { return nil,fmt.Errorf("DATABASE_URL is required") }
 db,err:=gorm.Open(postgres.Open(dsn), &gorm.Config{Logger:logger.Default.LogMode(logger.Silent)})
 if err!=nil{return nil,fmt.Errorf("database initialization failed")}
 pool,err:=db.DB();if err!=nil{return nil,err}
 pool.SetMaxOpenConns(25);pool.SetMaxIdleConns(5);pool.SetConnMaxLifetime(30*time.Minute);pool.SetConnMaxIdleTime(5*time.Minute)
 ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second);defer cancel()
 if err=pool.PingContext(ctx);err!=nil{pool.Close();return nil,fmt.Errorf("database connectivity failed")}
 return db,nil
}
