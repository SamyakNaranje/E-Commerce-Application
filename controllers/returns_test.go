package controllers
import "testing"
func TestRefundAmount(t *testing.T){
 for _,v:=range []struct{line,subtotal,discount,want int64}{
 {10000,10000,0,10000},{5000,10000,1000,4500},{333,1000,100,299},{100,100,100,0},{100,0,0,0},{100,100,101,0},{101,100,0,0},{-1,100,0,0},
 }{if got:=refundAmount(v.line,v.subtotal,v.discount);got!=v.want{t.Errorf("refundAmount(%d,%d,%d)=%d want %d",v.line,v.subtotal,v.discount,got,v.want)}}
}
