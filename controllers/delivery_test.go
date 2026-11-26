package controllers
import "testing"
func TestDeliveryPIN(t *testing.T){
 for _,v:=range []struct{s string;ok bool}{{"431001",true},{"411057",true},{"000001",false},{"12345",false},{"1234567",false},{"ABC123",false}}{
 if pinPattern.MatchString(v.s)!=v.ok{t.Errorf("PIN %q",v.s)}
 }
}
func TestDeliveryPhone(t *testing.T){
 for _,v:=range []struct{s string;ok bool}{{"9876543210",true},{"6123456789",true},{"5123456789",false},{"123",false},{"abcdefghij",false}}{
 if phonePattern.MatchString(v.s)!=v.ok{t.Errorf("phone %q",v.s)}
 }
}
