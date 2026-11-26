package services

import (
 "context"
 "errors"
 "gorm.io/gorm"
 "gorm.io/gorm/clause"
 "github.com/SamyakNaranje/E-Commerce-Application/models"
)
var ErrInvalidStock = errors.New("invalid stock adjustment")
type Inventory struct { DB *gorm.DB }
func StockStatus(q int64) string {
 if q<=0{return "out_of_stock"}
 if q<10{return "low_stock"}
 return "in_stock"
}
// AdjustTx must be called inside a transaction. All stock writers must use the same row lock.
func AdjustTx(tx *gorm.DB,id uint64,delta int64,reason string)(models.Inventory,error){
 var stock models.Inventory
 if delta==0 || delta < -1000000000 || delta>1000000000 || reason=="" {return stock,ErrInvalidStock}
 if err:=tx.Clauses(clause.Locking{Strength:"UPDATE"}).First(&stock,"product_id = ?",id).Error;err!=nil{return stock,err}
 if stock.Quantity+delta<0 || stock.Quantity+delta>1000000000{return stock,ErrInvalidStock}
 stock.Quantity+=delta
 if err:=tx.Model(&stock).Update("quantity",stock.Quantity).Error;err!=nil{return stock,err}
 if err:=tx.Create(&models.StockMovement{ProductID:id,Delta:delta,Reason:reason}).Error;err!=nil{return stock,err}
 return stock,nil
}
func (s Inventory) Adjust(ctx context.Context,id uint64,delta int64,reason string,actor uint64)(models.Inventory,error){
 var stock models.Inventory
 err:=s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB)error{
 var err error
 stock,err=AdjustTx(tx,id,delta,reason)
 if err!=nil{return err}
 return tx.Create(&models.AuditLog{UserID:actor,Action:"stock_adjustment",Resource:reason}).Error
 })
 return stock,err
}
