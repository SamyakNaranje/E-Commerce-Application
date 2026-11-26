package config

import (
 "os"
 "fmt"
 "gorm.io/driver/postgres"
 "gorm.io/gorm"
)

func Database() (*gorm.DB,error) {
 dsn := os.Getenv("DATABASE_URL")
 if dsn == "" { return nil,fmt.Errorf("DATABASE_URL is required") }
 return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
