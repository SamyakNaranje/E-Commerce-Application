package controllers
import (
 "encoding/csv"
 "strconv"
 "time"
 "github.com/gin-gonic/gin"
 "gorm.io/gorm"
 "github.com/SamyakNaranje/E-Commerce-Application/models"
)
type Dashboard struct{DB *gorm.DB}
func(h Dashboard) Summary(c *gin.Context){
 db:=h.DB.WithContext(c.Request.Context())
 out:=gin.H{}
 for _,v:=range []struct{key string;model interface{};where string}{
 {"customers",&models.User{},"role = 'customer'"},
 {"orders",&models.Order{},"1=1"},
 {"pending_orders",&models.Order{},"status = 'pending'"},
 {"pending_payments",&models.Payment{},"status IN ('pending','submitted')"},
 {"low_stock",&models.Inventory{},"quantity BETWEEN 1 AND 9"},
 {"out_of_stock",&models.Inventory{},"quantity = 0"},
 {"pending_returns",&models.ReturnRequest{},"status = 'pending'"},
 {"pending_refunds",&models.Refund{},"status = 'pending'"},
 }{
 var n int64
 if e:=db.Model(v.model).Where(v.where).Count(&n).Error;e!=nil{catalogError(c,e);return};out[v.key]=n
 }
 var gross,refund int64
 if e:=db.Model(&models.Payment{}).Where("status = ?","paid").Select("COALESCE(SUM(amount_paise),0)").Scan(&gross).Error;e!=nil{catalogError(c,e);return}
 if e:=db.Model(&models.Refund{}).Where("status = ?","completed").Select("COALESCE(SUM(amount_paise),0)").Scan(&refund).Error;e!=nil{catalogError(c,e);return}
 out["gross_paid_paise"]=gross;out["completed_refunds_paise"]=refund;out["net_receipts_paise"]=gross-refund
 c.JSON(200,out)
}
type salesRow struct{Day string `json:"day"`;Orders int64 `json:"orders"`;GrossPaise int64 `json:"gross_paise"`}
func(h Dashboard) sales(c *gin.Context)([]salesRow,bool){
 days,e:=strconv.Atoi(c.DefaultQuery("days","30"))
 if e!=nil||days<1||days>366{c.JSON(400,gin.H{"error":"days must be 1 to 366"});return nil,false}
 loc:=time.FixedZone("IST",19800)
 now:=time.Now().In(loc)
 end:=time.Date(now.Year(),now.Month(),now.Day(),0,0,0,0,loc).AddDate(0,0,1)
 start:=end.AddDate(0,0,-days)
 rows:=[]salesRow{}
 // Payment updated_at is the current paid-status timestamp, not a settlement ledger.
 err:=h.DB.WithContext(c.Request.Context()).Raw("SELECT to_char(updated_at AT TIME ZONE 'Asia/Kolkata','YYYY-MM-DD') AS day, COUNT(*) AS orders, COALESCE(SUM(amount_paise),0) AS gross_paise FROM payments WHERE status = ? AND updated_at >= ? AND updated_at < ? GROUP BY 1 ORDER BY 1","paid",start,end).Scan(&rows).Error
 if err!=nil{catalogError(c,err);return nil,false}
 byDay:=map[string]salesRow{}
 for _,r:=range rows{byDay[r.Day]=r}
 result:=[]salesRow{}
 for d:=start;d.Before(end);d=d.AddDate(0,0,1){key:=d.Format("2006-01-02");r:=byDay[key];r.Day=key;result=append(result,r)}
 return result,true
}
func(h Dashboard) Sales(c *gin.Context){if rows,ok:=h.sales(c);ok{c.JSON(200,rows)}}
func(h Dashboard) CSV(c *gin.Context){
 rows,ok:=h.sales(c);if !ok{return}
 c.Header("Content-Disposition","attachment; filename=sales.csv")
 c.Header("Content-Type","text/csv; charset=utf-8")
 w:=csv.NewWriter(c.Writer)
 _=w.Write([]string{"date","paid_orders","gross_paise"})
 for _,r:=range rows{_=w.Write([]string{r.Day,strconv.FormatInt(r.Orders,10),strconv.FormatInt(r.GrossPaise,10)})};w.Flush()
}
func(h Dashboard) BestSellers(c *gin.Context){
 rows:=[]struct{ProductID uint64 `json:"product_id"`;ProductName string `json:"product_name"`;Quantity int64 `json:"quantity"`}{}
 e:=h.DB.Raw("SELECT oi.product_id, MAX(oi.product_name) AS product_name, SUM(oi.quantity) AS quantity FROM order_items oi JOIN orders o ON o.id=oi.order_id WHERE o.status = ? GROUP BY oi.product_id ORDER BY quantity DESC, oi.product_id LIMIT 20","delivered").Scan(&rows).Error
 if e!=nil{catalogError(c,e);return};c.JSON(200,rows)
}
func(h Dashboard) Customers(c *gin.Context){
 page,e:=strconv.Atoi(c.DefaultQuery("page","1"))
 if e!=nil||page<1||page>100000{c.JSON(400,gin.H{"error":"invalid page"});return}
 rows:=[]models.User{}
 if e:=h.DB.Where("role = ?","customer").Order("id DESC").Limit(50).Offset((page-1)*50).Find(&rows).Error;e!=nil{catalogError(c,e);return};c.JSON(200,rows)
}
