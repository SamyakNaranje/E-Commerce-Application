package middleware

import (
 "net/http"
 "context"
 "net/url"
 "strconv"
 "context"
 "log"
 "github.com/SamyakNaranje/E-Commerce-Application/models"
 "os"
 "strings"
 "sync"
 "time"
 "github.com/gin-gonic/gin"
)
func Security() gin.HandlerFunc {
 return func(c *gin.Context){
 c.Header("X-Content-Type-Options","nosniff")
 c.Header("X-Frame-Options","DENY")
 c.Header("Referrer-Policy","strict-origin-when-cross-origin")
 c.Header("Permissions-Policy","camera=(), microphone=(), geolocation=()")
 c.Header("Content-Security-Policy","default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'; form-action 'self'")
 if os.Getenv("COOKIE_SECURE")!="false"{c.Header("Strict-Transport-Security","max-age=31536000")}
 if strings.HasPrefix(c.Request.URL.Path,"/api/"){
 c.Header("Cache-Control","no-store")
 if len(c.Request.URL.RawQuery)>2048 {c.AbortWithStatusJSON(400,gin.H{"error":"query too long"});return}
 for _,p:=range c.Params{if p.Key=="id"||p.Key=="image"{n,e:=strconv.ParseUint(p.Value,10,63);if e!=nil||n==0{c.AbortWithStatusJSON(400,gin.H{"error":"invalid identifier"});return}}}
 ctx,cancel:=context.WithTimeout(c.Request.Context(),15*time.Second);defer cancel();c.Request=c.Request.WithContext(ctx)
 if origin:=c.GetHeader("Origin");origin!="" && os.Getenv("APP_ORIGIN")!="" && origin!=strings.TrimRight(os.Getenv("APP_ORIGIN"),"/"){c.AbortWithStatusJSON(403,gin.H{"error":"origin rejected"});return}
 if c.Request.ContentLength>1<<20{c.AbortWithStatusJSON(413,gin.H{"error":"request too large"});return}
 ctx,cancel:=context.WithTimeout(c.Request.Context(),15*time.Second);defer cancel()
 c.Request=c.Request.WithContext(ctx)
 if len(c.Request.URL.RawQuery)>4096{c.AbortWithStatusJSON(414,gin.H{"error":"query too long"});return}
 if c.Request.ContentLength>1<<20{c.AbortWithStatusJSON(413,gin.H{"error":"request too large"});return}
 if origin:=c.GetHeader("Origin");origin!=""&&os.Getenv("APP_ORIGIN")!=""&&origin!=strings.TrimRight(os.Getenv("APP_ORIGIN"),"/"){c.AbortWithStatusJSON(403,gin.H{"error":"origin rejected"});return}
 c.Request.Body=http.MaxBytesReader(c.Writer,c.Request.Body,1<<20)
 switch c.Request.Method{
 case "GET","HEAD","OPTIONS":
 default:
 if c.GetHeader("Sec-Fetch-Site")=="cross-site"{c.AbortWithStatusJSON(403,gin.H{"error":"cross-site request rejected"});return}
 origin:=c.GetHeader("Origin")
 if origin!=""{
 u,e:=url.Parse(origin)
 expected:=strings.TrimRight(os.Getenv("APP_ORIGIN"),"/")
 valid:=e==nil && u.User==nil && u.Path=="" && u.RawQuery=="" && u.Fragment=="" && (u.Scheme=="https"||u.Scheme=="http")
 if expected!=""{valid=valid&&origin==expected}else{scheme:="http";if c.Request.TLS!=nil{scheme="https"};valid=valid&&u.Scheme==scheme&&strings.EqualFold(u.Host,c.Request.Host)}
 if !valid{c.AbortWithStatusJSON(403,gin.H{"error":"origin rejected"});return}
 }
 if origin=="" && c.GetHeader("Sec-Fetch-Site")!="same-origin"{c.AbortWithStatusJSON(403,gin.H{"error":"origin required"});return}
 }
 }
 c.Next()
 }
}
type authBucket struct{Count int;Until time.Time}
func AuthLimit() gin.HandlerFunc{return RateLimit(10)}
func RateLimit(max int) gin.HandlerFunc{return RateLimit(10)}
func RateLimit(max int) gin.HandlerFunc{
 var mu sync.Mutex
 buckets:=map[string]authBucket{}
 var sweep time.Time
 var sweep time.Time
 return func(c *gin.Context){
 now:=time.Now();key:=c.ClientIP()
 mu.Lock()
 if now.After(sweep){if now.After(sweep){for k,v:=range buckets{if !now.Before(v.Until){delete(buckets,k)}};sweep=now.Add(time.Minute)};sweep=now.Add(time.Minute)}
 b,exists:=buckets[key]
 if !exists && len(buckets)>=10000{mu.Unlock();c.AbortWithStatusJSON(429,gin.H{"error":"try again later"});return}
 if !exists||!now.Before(b.Until){b=authBucket{Until:now.Add(time.Minute)}}
 if b.Count>=max{mu.Unlock();c.Header("Retry-After","60");c.AbortWithStatusJSON(429,gin.H{"error":"too many attempts"});return}
 b.Count++;buckets[key]=b;mu.Unlock();c.Next()
 }
}

func Audit() gin.HandlerFunc{return func(c *gin.Context){
 c.Next()
 if !strings.HasPrefix(c.Request.URL.Path,"/api/"){return}
 if c.Request.Method=="GET"||c.Request.Method=="HEAD"||c.Request.Method=="OPTIONS"{return}
 var uid uint64
 if v,ok:=c.Get("user");ok{if u,ok:=v.(models.User);ok{uid=u.ID}}
 log.Printf("security_event method=%q route=%q actor=%d status=%d ip=%q",c.Request.Method,c.FullPath(),uid,c.Writer.Status(),c.ClientIP())
}}
