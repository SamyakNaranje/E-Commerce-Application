package middleware

import (
 "net/http"
 "net/url"
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
 if _,e:=c.Cookie("session");e==nil && origin=="" && c.GetHeader("Sec-Fetch-Site")!="same-origin"{c.AbortWithStatusJSON(403,gin.H{"error":"origin required"});return}
 }
 }
 c.Next()
 }
}
type authBucket struct{Count int;Until time.Time}
func AuthLimit() gin.HandlerFunc{
 var mu sync.Mutex
 buckets:=map[string]authBucket{}
 return func(c *gin.Context){
 now:=time.Now();key:=c.ClientIP()
 mu.Lock()
 for k,v:=range buckets{if !now.Before(v.Until){delete(buckets,k)}}
 b,exists:=buckets[key]
 if !exists && len(buckets)>=10000{mu.Unlock();c.AbortWithStatusJSON(429,gin.H{"error":"try again later"});return}
 if !exists{b=authBucket{Until:now.Add(time.Minute)}}
 if b.Count>=10{mu.Unlock();c.Header("Retry-After","60");c.AbortWithStatusJSON(429,gin.H{"error":"too many attempts"});return}
 b.Count++;buckets[key]=b;mu.Unlock();c.Next()
 }
}
