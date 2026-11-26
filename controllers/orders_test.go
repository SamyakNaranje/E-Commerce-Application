package controllers
import "testing"
func TestOrderTransitions(t *testing.T){
 cases:=[]struct{from,to,method,payment string;admin,want bool}{
 {"pending","confirmed","cod","pending",true,true},
 {"pending","confirmed","upi","pending",true,false},
 {"pending","confirmed","cod","pending",false,false},
 {"pending","cancelled","upi","pending",false,true},
 {"pending","cancelled","upi","submitted",false,false},
 {"confirmed","cancelled","upi","paid",true,false},
 {"confirmed","packed","upi","paid",true,true},
 {"confirmed","packed","upi","pending",true,false},
 {"packed","shipped","cod","pending",true,true},
 {"shipped","delivered","cod","pending",true,true},
 {"shipped","delivered","upi","pending",true,false},
 {"pending","delivered","cod","pending",true,false},
 {"delivered","cancelled","cod","paid",true,false},
 {"cancelled","confirmed","cod","pending",true,false},
 }
 for _,c:=range cases{if got:=allowedOrderTransition(c.from,c.to,c.method,c.payment,c.admin);got!=c.want{t.Errorf("%+v: got %v",c,got)}}
}
