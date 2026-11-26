package services

import (
 "errors"
 "strings"
 "net/mail"
 "time"
 "os"
 "github.com/golang-jwt/jwt/v5"
 "golang.org/x/crypto/bcrypt"
 "gorm.io/gorm"
 "github.com/SamyakNaranje/E-Commerce-Application/models"
)

type Auth struct { DB *gorm.DB }
type Claims struct { UserID uint64 `json:"uid"`; jwt.RegisteredClaims }
func secret() ([]byte,error) {
 s:=os.Getenv("JWT_SECRET")
 if len(s)<32 { return nil,errors.New("JWT_SECRET must be at least 32 bytes") }
 return []byte(s),nil
}
func (a Auth) Register(name,email,password string) error {
 email=strings.ToLower(strings.TrimSpace(email))
 name=strings.TrimSpace(name)
 addr,err:=mail.ParseAddress(email)
 if err!=nil || addr.Address!=email || len(email)>254 || len(name)==0 || len(name)>120 || len(password)<12 || len(password)>72 { return errors.New("invalid registration") }
 hash,err:=bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)
 if err!=nil{return err}
 return a.DB.Create(&models.User{Name:name,Email:email,PasswordHash:string(hash),Role:"customer",Active:true}).Error
}
func (a Auth) Login(email,password string)(string,error){
 var u models.User
 if len(password)>72 || a.DB.Where("email = ?",strings.ToLower(strings.TrimSpace(email))).First(&u).Error!=nil || !u.Active || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash),[]byte(password))!=nil {return "",errors.New("invalid credentials")}
 key,err:=secret();if err!=nil{return "",err}
 now:=time.Now()
 return jwt.NewWithClaims(jwt.SigningMethodHS256,Claims{UserID:u.ID,RegisteredClaims:jwt.RegisteredClaims{Issuer:"ecommerce",IssuedAt:jwt.NewNumericDate(now),ExpiresAt:jwt.NewNumericDate(now.Add(30*time.Minute))}}).SignedString(key)
}
func (a Auth) Verify(raw string)(models.User,error){
 var u models.User
 key,err:=secret();if err!=nil{return u,err}
 claims:=&Claims{}
 token,err:=jwt.ParseWithClaims(raw,claims,func(t *jwt.Token)(interface{},error){return key,nil},jwt.WithValidMethods([]string{"HS256"}),jwt.WithIssuer("ecommerce"),jwt.WithExpirationRequired())
 if err!=nil || !token.Valid || claims.UserID==0{return u,errors.New("invalid token")}
 err=a.DB.First(&u,claims.UserID).Error
 if err!=nil || !u.Active{return u,errors.New("inactive account")}
 return u,nil
}
