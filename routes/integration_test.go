package routes
import(
 "os"
 "testing"
 "net/http/httptest"
 "strings"
 "github.com/gin-gonic/gin"
 "gorm.io/gorm"
 "gorm.io/driver/postgres"
 "github.com/SamyakNaranje/E-Commerce-Application/models"
)
func TestIntegrationAuth(t *testing.T){
 dsn:=os.Getenv("TEST_DATABASE_URL");if dsn==""{t.Skip("dedicated test database required")}
 t.Setenv("JWT_SECRET","integration-test-secret-not-for-production")
 t.Setenv("COOKIE_SECURE","false")
 db,e:=gorm.Open(postgres.Open(dsn),&gorm.Config{});if e!=nil{t.Fatal(e)}
 sqlDB,e:=db.DB();if e!=nil{t.Fatal(e)};defer sqlDB.Close()
 if e=db.AutoMigrate(&models.User{});e!=nil{t.Fatal(e)}
 email:="integration@example.invalid"
 if e=db.Where("email = ?",email).Delete(&models.User{}).Error;e!=nil{t.Fatal(e)}
 defer db.Where("email = ?",email).Delete(&models.User{})
 r:=gin.New();Register(r,db)
 call:=func(method,path,body string,cookie string)*httptest.ResponseRecorder{
 req:=httptest.NewRequest(method,path,strings.NewReader(body));req.Header.Set("Content-Type","application/json");req.Header.Set("Origin","http://example.com");if cookie!=""{req.Header.Set("Cookie",cookie)}
 w:=httptest.NewRecorder();r.ServeHTTP(w,req);return w
 }
 if w:=call("GET","/api/cart","","");w.Code!=401{t.Fatalf("anonymous cart: %d",w.Code)}
 if w:=call("POST","/api/auth/register",`{"name":"Integration","email":"integration@example.invalid","password":"TestPassword123!"}`,"");w.Code!=201{t.Fatalf("register: %d %s",w.Code,w.Body)}
 w:=call("POST","/api/auth/login",`{"email":"integration@example.invalid","password":"TestPassword123!"}`,"")
 if w.Code!=200{t.Fatalf("login: %d %s",w.Code,w.Body)}
 cookies:=w.Result().Cookies();if len(cookies)==0{t.Fatal("missing session cookie")}
 cookie:=cookies[0].Name+"="+cookies[0].Value
 if w:=call("GET","/api/auth/me","",cookie);w.Code!=200{t.Fatalf("me: %d",w.Code)}
 if w:=call("GET","/api/admin/dashboard","",cookie);w.Code!=403{t.Fatalf("customer admin access: %d",w.Code)}
}
