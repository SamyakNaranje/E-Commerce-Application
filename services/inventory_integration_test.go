package services

import (
 "errors"
 "fmt"
 "os"
 "sync"
 "testing"
 "time"
 "gorm.io/driver/postgres"
 "gorm.io/gorm"
 "github.com/SamyakNaranje/E-Commerce-Application/models"
)
func TestInventoryConcurrencyAndRollback(t *testing.T){
 dsn:=os.Getenv("TEST_DATABASE_URL")
 if dsn==""{t.Skip("dedicated test database required")}
 db,e:=gorm.Open(postgres.Open(dsn),&gorm.Config{});if e!=nil{t.Fatal(e)}
 sqlDB,e:=db.DB();if e!=nil{t.Fatal(e)};defer sqlDB.Close()
 if e=db.AutoMigrate(&models.Category{},&models.Product{},&models.Inventory{},&models.StockMovement{});e!=nil{t.Fatal(e)}
 suffix:=fmt.Sprint(time.Now().UnixNano())
 cat:=models.Category{Name:"Concurrency "+suffix}
 if e=db.Create(&cat).Error;e!=nil{t.Fatal(e)}
 p:=models.Product{CategoryID:cat.ID,Name:"Concurrent stock test",SKU:"concurrency-"+suffix,PricePaise:100}
 if e=db.Create(&p).Error;e!=nil{t.Fatal(e)}
 if e=db.Create(&models.Inventory{ProductID:p.ID,Quantity:1}).Error;e!=nil{t.Fatal(e)}
 defer func(){db.Where("product_id = ?",p.ID).Delete(&models.StockMovement{});db.Where("product_id = ?",p.ID).Delete(&models.Inventory{});db.Delete(&p);db.Delete(&cat)}()
 const workers=20
 start:=make(chan struct{})
 results:=make(chan error,workers)
 var wg sync.WaitGroup
 for i:=0;i<workers;i++{wg.Add(1);go func(){defer wg.Done();<-start;results<-db.Transaction(func(tx *gorm.DB)error{_,err:=AdjustTx(tx,p.ID,-1,"concurrent purchase");return err})}()}
 close(start);wg.Wait();close(results)
 success,rejected:=0,0
 for err:=range results{switch{case err==nil:success++;case errors.Is(err,ErrInvalidStock):rejected++;default:t.Errorf("unexpected transaction error: %v",err)}}
 if success!=1||rejected!=workers-1{t.Fatalf("success=%d rejected=%d",success,rejected)}
 check:=func(quantity,movements int64){t.Helper();var stock models.Inventory;var count int64;if e:=db.First(&stock,"product_id = ?",p.ID).Error;e!=nil{t.Fatal(e)};if e:=db.Model(&models.StockMovement{}).Where("product_id = ?",p.ID).Count(&count).Error;e!=nil{t.Fatal(e)};if stock.Quantity!=quantity||count!=movements{t.Fatalf("stock=%d movements=%d want %d/%d",stock.Quantity,count,quantity,movements)}}
 check(0,1)
 forced:=errors.New("forced rollback")
 e=db.Transaction(func(tx *gorm.DB)error{if _,err:=AdjustTx(tx,p.ID,5,"rollback test");err!=nil{return err};return forced})
 if !errors.Is(e,forced){t.Fatalf("expected rollback: %v",e)}
 check(0,1)
}
