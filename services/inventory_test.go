package services
import "testing"
func TestStockStatus(t *testing.T){
 for _,v:=range []struct{q int64;want string}{{0,"out_of_stock"},{1,"low_stock"},{9,"low_stock"},{10,"in_stock"},{100,"in_stock"}}{
 if got:=StockStatus(v.q);got!=v.want{t.Errorf("quantity %d: got %s want %s",v.q,got,v.want)}
 }
}
