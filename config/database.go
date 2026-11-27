package config
import("os";"fmt";"context";"time";"gorm.io/driver/postgres";"gorm.io/gorm";"gorm.io/gorm/logger")
func Database() (*gorm.DB,error){
 dsn:=os.Getenv("DATABASE_URL");if dsn==""{return nil,fmt.Errorf("DATABASE_URL is required")}
 db,e:=gorm.Open(postgres.Open(dsn),&gorm.Config{Logger:logger.Default.LogMode(logger.Silent)});if e!=nil{return nil,fmt.Errorf("database initialization failed")}
 pool,e:=db.DB();if e!=nil{return nil,e}
 pool.SetMaxOpenConns(25);pool.SetMaxIdleConns(5);pool.SetConnMaxLifetime(30*time.Minute);pool.SetConnMaxIdleTime(5*time.Minute)
 ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second);defer cancel()
 if e=pool.PingContext(ctx);e!=nil{pool.Close();return nil,fmt.Errorf("database unavailable")};return db,nil
}
