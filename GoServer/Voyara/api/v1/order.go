package v1

import "github.com/gogf/gf/v2/frame/g"

type CheckoutReq struct {
	g.Meta          `path:"/voyara/orders" method:"post" summary:"Checkout from cart" middleware:"auth"`
	ProductIDs      []int         `json:"productIds" v:"required|min-length:1"`
	IdempotencyKey  string        `json:"idempotencyKey"`
	ShippingAddress ShippingAddr  `json:"shippingAddress" v:"required"`
}

type ShippingAddr struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Country string `json:"country"`
	City    string `json:"city"`
	Street  string `json:"street"`
	ZipCode string `json:"zipCode"`
}

type GetOrdersReq struct {
	g.Meta `path:"/voyara/orders" method:"get" summary:"My orders" middleware:"auth"`
}

type ShipOrderReq struct {
	g.Meta         `path:"/voyara/orders/:id/ship" method:"put" summary:"Ship order" middleware:"auth"`
	ID             int    `json:"id" in:"path"`
	TrackingNumber string `json:"trackingNumber" v:"required"`
}

type OrderItemRes struct {
	ID              int           `json:"id"`
	OrderNo         string        `json:"orderNo"`
	BuyerID         int           `json:"buyerId"`
	SellerID        int           `json:"sellerId"`
	ItemCount       int           `json:"itemCount"`
	Amount          float64       `json:"amount"`
	Subtotal        float64       `json:"subtotal"`
	ShippingFee     float64       `json:"shippingFee"`
	Currency        string        `json:"currency"`
	PaymentStatus   string        `json:"paymentStatus"`
	ShippingStatus  string        `json:"shippingStatus"`
	TrackingNumber  string        `json:"trackingNumber"`
	ShippingAddress ShippingAddr  `json:"shippingAddress"`
	CreatedAt       string        `json:"createdAt"`
	Items           []OrderItem   `json:"items,omitempty"`
}

type OrderItem struct {
	ID        int     `json:"id"`
	ProductID int     `json:"productId"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Total     float64 `json:"total"`
	ImageURL  string  `json:"imageUrl"`
}

type GetOrdersRes []OrderItemRes

type CheckoutRes = OrderItemRes
