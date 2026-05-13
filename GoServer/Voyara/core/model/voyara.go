package model

import "database/sql"

type User struct {
	ID                 int            `json:"id"`
	Email              string         `json:"email"`
	PasswordHash       string         `json:"-"`
	PasswordHashMethod string         `json:"-"`
	Name               string         `json:"name"`
	Phone              string         `json:"phone"`
	Country            string         `json:"country"`
	PreferredLang      string         `json:"preferredLang"`
	EmailVerifiedAt    sql.NullTime   `json:"emailVerifiedAt"`
	Role               string         `json:"role"`
	LoginAttempts      int            `json:"-"`
	LockedUntil        sql.NullTime   `json:"-"`
	CreatedAt          string         `json:"createdAt"`
}

type Seller struct {
	ID          int            `json:"id"`
	UserID      int            `json:"userId"`
	ShopName    string         `json:"shopName"`
	Description sql.NullString `json:"description"`
	Verified    bool           `json:"verified"`
	Rating      float64        `json:"rating"`
	CreatedAt   string         `json:"createdAt"`
}

type Product struct {
	ID          int            `json:"id"`
	SellerID    int            `json:"sellerId"`
	SellerName  string         `json:"sellerName,omitempty"`
	ShopName    string         `json:"shopName,omitempty"`
	Title       string         `json:"title"`
	Description sql.NullString `json:"description"`
	Price       float64        `json:"price"`
	Currency    string         `json:"currency"`
	Category    string         `json:"category"`
	Condition   string         `json:"condition"`
	Images      sql.NullString `json:"images"`
	Status      string         `json:"status"`
	CreatedAt   string         `json:"createdAt"`
	UpdatedAt   string         `json:"updatedAt"`
}

type Category struct {
	ID       int           `json:"id"`
	Name     string        `json:"name"`
	ParentID sql.NullInt64 `json:"parentId"`
	Icon     string        `json:"icon"`
}

type CartItem struct {
	ID        int  `json:"id"`
	UserID    int  `json:"userId"`
	ProductID int  `json:"productId"`
	Quantity  int  `json:"quantity"`
	Selected  bool `json:"selected"`
}

type Order struct {
	ID              int            `json:"id"`
	OrderNo         string         `json:"orderNo"`
	BuyerID         int            `json:"buyerId"`
	SellerID        int            `json:"sellerId"`
	ProductID       int            `json:"productId"`
	ItemCount       int            `json:"itemCount"`
	Amount          float64        `json:"amount"`
	Subtotal        float64        `json:"subtotal"`
	Currency        string         `json:"currency"`
	PaymentStatus   string         `json:"paymentStatus"`
	ShippingStatus  string         `json:"shippingStatus"`
	TrackingNumber  string         `json:"trackingNumber"`
	ShippingAddress sql.NullString `json:"shippingAddress"`
	CreatedAt       string         `json:"createdAt"`
	Items           []OrderItem    `json:"items,omitempty"`
}

type OrderItem struct {
	ID        int     `json:"id"`
	OrderID   int     `json:"orderId"`
	ProductID int     `json:"productId"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Total     float64 `json:"total"`
	ImageURL  string  `json:"imageUrl"`
}

type VerificationCode struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Code      string `json:"-"`
	Purpose   string `json:"purpose"`
	ExpiresAt string `json:"expiresAt"`
	Used      bool   `json:"used"`
}
