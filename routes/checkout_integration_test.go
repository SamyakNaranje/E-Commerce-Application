package routes

import (
 "encoding/json"
 "fmt"
 "net/http/httptest"
 "os"
 "strings"
 "testing"
 "time"
 "github.com/gin-gonic/gin"
 "gorm.io/driver/postgres"
 "gorm.io/gorm"
 "github.com/SamyakNaranje/E-Commerce-Application/models"
)
func TestCheckoutLifecycle(t *testing.T){
 dsn:=os.Getenv("TEST_DATABASE_URL");if dsn==""{t.Skip("dedicated test database required")}
 t.Setenv("JWT_SECRET","integration-test-secret-not-for-production")
 t.Setenv("COOKIE_SECURE","false")
 t.Setenv("APP_ORIGIN","http://example.com")
 db,e:=gorm.Open(postgres.Open(dsn),&gorm.Config{});if e!=nil{t.Fatal(e)}
 sqlDB,e:=db.DB();if e!=nil{t.Fatal(e)};defer sqlDB.Close()
 if e=db.AutoMigrate(&models.User{},&models.Category{},&models.Product{},&models.Inventory{},&models.StockMovement{},&models.Address{},&models.CartItem{},&models.DeliveryArea{},&models.Coupon{},&models.Order{},&models.OrderItem{},&models.Payment{},&models.OrderEvent{},&models.AuditLog{},&models.ReturnRequest{},&models.Refund{});e!=nil{t.Fatal(e)}
 suffix:=fmt.Sprint(time.Now().UnixNano())
 email:="checkout-"+suffix+"@example.invalid"
 r:=gin.New();Register(r,db)
 cookie:=""
 call:=func(method,path,body string,want int)*httptest.ResponseRecorder{
 t.Helper()
 req:=httptest.NewRequest(method,path,strings.NewReader(body));req.Header.Set("Content-Type","application/json");req.Header.Set("Origin","http://example.com");if cookie!=""{req.Header.Set("Cookie",cookie)}
 w:=httptest.NewRecorder();r.ServeHTTP(w,req);if w.Code!=want{t.Fatalf("%s %s: got %d want %d: %s",method,path,w.Code,want,w.Body)};return w
 }
 credentials:=fmt.Sprintf(`{"email":%q,"password":"TestPassword123!","name":"Checkout test"}`,email)
 call("POST","/api/auth/register",credentials,201)
 w:=call("POST","/api/auth/login",credentials,200)
 cs:=w.Result().Cookies();if len(cs)==0{t.Fatal("missing cookie")};cookie=cs[0].Name+"="+cs[0].Value
 create:=func(v interface{}){t.Helper();if e:=db.Create(v).Error;e!=nil{t.Fatal(e)}}
 cat:=models.Category{Name:"Test "+suffix};create(&cat)
 p:=models.Product{Name:"Test product",SKU:suffix,CategoryID:cat.ID,PricePaise:10000,Active:true};create(&p)
 create(&models.Inventory{ProductID:p.ID,Quantity:3})
 area:=models.DeliveryArea{PostalCode:"999999",ChargePaise:1000,Active:true}
 if e:=db.Where("postal_code = ?",area.PostalCode).Assign(area).FirstOrCreate(&area).Error;e!=nil{t.Fatal(e)}
 w=call("POST","/api/addresses",`{"recipient":"Test","phone":"9876543210","line":"Test street","city":"Test city","postal_code":"999999"}`,201)
 var a models.Address;if e=json.Unmarshal(w.Body.Bytes(),&a);e!=nil{t.Fatal(e)}
 path:=fmt.Sprintf("/api/cart/%d",p.ID)
 call("PUT",path,`{"quantity":4}`,409)
 call("PUT",path,`{"quantity":2}`,204)
 body:=fmt.Sprintf(`{"address_id":%d,"method":"cod","idempotency_key":%q}`,a.ID,"checkout-"+suffix)
 w=call("POST","/api/checkout",body,200)
 var o models.Order;if e=json.Unmarshal(w.Body.Bytes(),&o);e!=nil{t.Fatal(e)}
 if o.TotalPaise!=21000||o.ID==0{t.Fatalf("unexpected order: %+v",o)}
 w=call("POST","/api/checkout",body,200)
 var duplicate models.Order;if e=json.Unmarshal(w.Body.Bytes(),&duplicate);e!=nil||duplicate.ID!=o.ID{t.Fatal("idempotency failed")}
 stock:=func(want int64){t.Helper();var inv models.Inventory;if e:=db.First(&inv,"product_id = ?",p.ID).Error;e!=nil||inv.Quantity!=want{t.Fatalf("stock: %d want %d error %v",inv.Quantity,want,e)}}
 stock(1)
 cancel:=fmt.Sprintf("/api/orders/%d/cancel",o.ID)
 call("POST",cancel,"",200);stock(3)
 call("POST",cancel,"",200);stock(3)
 customerCookie:=cookie
 t.Setenv("UPI_ID","test@example")
 t.Setenv("UPI_PAYEE_NAME","Test store")
 call("PUT",path,`{"quantity":1}`,204)
 body=fmt.Sprintf(`{"address_id":%d,"method":"upi","idempotency_key":%q}`,a.ID,"payment-"+suffix)
 w=call("POST","/api/checkout",body,200)
 var paidOrder models.Order
 if e=json.Unmarshal(w.Body.Bytes(),&paidOrder);e!=nil{t.Fatal(e)}
 stock(2)
 orderPath:=fmt.Sprintf("/api/orders/%d",paidOrder.ID)
 w=call("GET",orderPath+"/payment","",200)
 if !strings.Contains(w.Body.String(),"upi://pay?"){t.Fatal("missing UPI URI")}
 call("POST",orderPath+"/payment/reference",`{"reference":"TEST123456"}`,200)
 call("POST",orderPath+"/payment/reference",`{"reference":"TEST123456"}`,409)
 adminEmail:="admin-"+suffix+"@example.invalid"
 adminCredentials:=fmt.Sprintf(`{"email":%q,"password":"TestPassword123!","name":"Test admin"}`,adminEmail)
 call("POST","/api/auth/register",adminCredentials,201)
 if e=db.Model(&models.User{}).Where("email = ?",adminEmail).Update("role","admin").Error;e!=nil{t.Fatal(e)}
 w=call("POST","/api/auth/login",adminCredentials,200)
 cs=w.Result().Cookies();if len(cs)==0{t.Fatal("missing admin cookie")};adminCookie:=cs[0].Name+"="+cs[0].Value
 call("POST",fmt.Sprintf("/api/admin/orders/%d/payment/verify",paidOrder.ID),`{"decision":"approve","note":"Test verification"}`,403)
 cookie=adminCookie
 adminPath:=fmt.Sprintf("/api/admin/orders/%d",paidOrder.ID)
 call("POST",adminPath+"/payment/verify",`{"decision":"approve","note":"Test verification"}`,200)
 for _,status:=range []string{"packed","shipped","delivered"}{call("PUT",adminPath+"/status",fmt.Sprintf(`{"status":%q}`,status),200)}
 var item models.OrderItem
 if e=db.Where("order_id = ?",paidOrder.ID).First(&item).Error;e!=nil{t.Fatal(e)}
 cookie=customerCookie
 request:=fmt.Sprintf(`{"order_item_id":%d,"reason":"Test return"}`,item.ID)
 w=call("POST","/api/returns",request,201)
 var ret models.ReturnRequest;if e=json.Unmarshal(w.Body.Bytes(),&ret);e!=nil{t.Fatal(e)}
 call("POST","/api/returns",request,409)
 cookie=adminCookie
 retPath:=fmt.Sprintf("/api/admin/returns/%d/status",ret.ID)
 call("PUT",retPath,`{"status":"approved","note":"Approved"}`,204)
 call("PUT",retPath,`{"status":"received","note":"Received","restock":true}`,204);stock(3)
 call("PUT",retPath,`{"status":"received","note":"Received","restock":true}`,204);stock(3)
 call("PUT",retPath,`{"status":"refunded","note":"Refunded","reference":"TESTREFUND"}`,204)
 call("PUT",retPath,`{"status":"refunded","note":"Refunded","reference":"TESTREFUND"}`,204)
 var refunds []models.Refund
 if e=db.Where("return_request_id = ?",ret.ID).Find(&refunds).Error;e!=nil{t.Fatal(e)}
 if len(refunds)!=1||refunds[0].Status!="completed"||refunds[0].AmountPaise!=10000{t.Fatalf("unexpected refunds: %+v",refunds)}
}

