package middleware
import("net/http/httptest";"testing";"github.com/gin-gonic/gin")
func TestCrossOrigin(t *testing.T){
 t.Setenv("APP_ORIGIN","https://shop.example")
 r:=gin.New();r.Use(Security());r.POST("/api/test",func(c *gin.Context){c.Status(204)})
 for _,v:=range []struct{origin string;want int}{{"https://evil.example",403},{"https://shop.example",204}}{
 req:=httptest.NewRequest("POST","/api/test",nil);req.Header.Set("Origin",v.origin);w:=httptest.NewRecorder();r.ServeHTTP(w,req);if w.Code!=v.want{t.Errorf("origin %s: got %d want %d",v.origin,w.Code,v.want)}
 }
}
func TestAuthLimit(t *testing.T){
 r:=gin.New();r.SetTrustedProxies(nil);r.POST("/login",AuthLimit(),func(c *gin.Context){c.Status(204)})
 for i:=0;i<11;i++{w:=httptest.NewRecorder();r.ServeHTTP(w,httptest.NewRequest("POST","/login",nil));want:=204;if i==10{want=429};if w.Code!=want{t.Fatalf("attempt %d: %d",i,w.Code)}}
}
