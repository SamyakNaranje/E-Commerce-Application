package models

import "time"

type Base struct {
 ID uint64 `gorm:"primaryKey" json:"id"`
 CreatedAt time.Time `json:"created_at"`
 UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
 Base
 Name string `gorm:"not null" json:"name"`
 Email string `gorm:"uniqueIndex;not null" json:"email"`
 PasswordHash string `gorm:"not null" json:"-"`
 Role string `gorm:"not null;default:customer;check:role IN ('admin','customer')" json:"role"`
 Active bool `gorm:"not null;default:true" json:"active"`
}
type Category struct {
 Base
 Name string `gorm:"uniqueIndex;not null" json:"name"`
}
type Product struct {
 Base
 CategoryID uint64 `gorm:"not null;index" json:"category_id"`
 Category Category `json:"category,omitempty"`
 Name string `gorm:"not null" json:"name"`
 Description string `json:"description"`
 SKU string `gorm:"uniqueIndex;not null" json:"sku"`
 PricePaise int64 `gorm:"not null;check:price_paise >= 0" json:"price_paise"`
 Active bool `gorm:"not null;default:true" json:"active"`
}
type ProductImage struct {
 Base
 ProductID uint64 `gorm:"not null;index"`
 Product Product `json:"-"`
 URL string `gorm:"not null"`
}
type Inventory struct {
 ProductID uint64 `gorm:"primaryKey;autoIncrement:false" json:"product_id"`
 Product Product `json:"-"`
 Quantity int64 `gorm:"not null;check:quantity >= 0" json:"quantity"`
 UpdatedAt time.Time
}
type StockMovement struct {
 Base
 ProductID uint64 `gorm:"not null;index"`
 Product Product
 Delta int64
 Reason string
}
type Address struct {
 Base
 UserID uint64 `gorm:"not null;index"`
 User User `json:"-"`
 Recipient string
 Phone string
 Line string
 City string
 PostalCode string
}
type CartItem struct {
 Base
 UserID uint64 `gorm:"not null;uniqueIndex:cart_product"`
 User User `json:"-"`
 ProductID uint64 `gorm:"not null;uniqueIndex:cart_product"`
 Product Product
 Quantity int64 `gorm:"not null;check:quantity > 0"`
}
type WishlistItem struct {
 Base
 UserID uint64 `gorm:"not null;uniqueIndex:wishlist_product"`
 User User `json:"-"`
 ProductID uint64 `gorm:"not null;uniqueIndex:wishlist_product"`
 Product Product
}
type Coupon struct {
 Base
 Code string `gorm:"uniqueIndex;not null"`
 DiscountPercent int64 `gorm:"check:discount_percent BETWEEN 1 AND 100"`
 ExpiresAt time.Time
 Active bool
}
type Order struct {
 Base
 UserID uint64 `gorm:"not null;index;uniqueIndex:checkout_user_key"`
 CheckoutKey string `gorm:"not null;uniqueIndex:checkout_user_key" json:"-"`
 SubtotalPaise int64
 DiscountPaise int64
 DeliveryChargePaise int64
 CouponCode string
 User User `json:"-"`
 Status string `gorm:"not null;default:pending"`
 TotalPaise int64 `gorm:"check:total_paise >= 0"`
 DeliveryAddress string `gorm:"not null"`
 Items []OrderItem
}
type OrderItem struct {
 Base
 OrderID uint64 `gorm:"not null;index"`
 ProductID uint64 `gorm:"not null"`
 Product Product
 ProductName string
 Quantity int64 `gorm:"check:quantity > 0"`
 UnitPricePaise int64 `gorm:"check:unit_price_paise >= 0"`
}
type Payment struct {
 Base
 OrderID uint64 `gorm:"not null;uniqueIndex"`
 Order Order `json:"-"`
 Method string
 Status string `gorm:"default:pending"`
 AmountPaise int64 `gorm:"check:amount_paise >= 0"`
 Reference string
}
type ReturnRequest struct {
 Base
 OrderItemID uint64 `gorm:"not null;index"`
 OrderItem OrderItem
 Reason string
 Status string `gorm:"default:pending"`
}
type Refund struct {
 Base
 ReturnRequestID uint64 `gorm:"not null;index"`
 ReturnRequest ReturnRequest
 AmountPaise int64 `gorm:"check:amount_paise >= 0"`
 Status string
}
type Review struct {
 Base
 UserID uint64 `gorm:"not null;uniqueIndex:review_product"`
 User User `json:"-"`
 ProductID uint64 `gorm:"not null;uniqueIndex:review_product"`
 Product Product
 Rating int `gorm:"check:rating BETWEEN 1 AND 5"`
 Comment string
}
type DeliveryArea struct {
 Base
 PostalCode string `gorm:"uniqueIndex;not null"`
 ChargePaise int64 `gorm:"check:charge_paise >= 0"`
 Active bool
}
type AuditLog struct {
 Base
 UserID uint64 `gorm:"index"`
 Action string
 Resource string
}

type OrderEvent struct {
 Base
 OrderID uint64 `gorm:"not null;index"`
 Order Order `json:"-"`
 FromStatus string
 ToStatus string
 ActorID uint64
}
