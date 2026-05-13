package service

type PaymentMethod string

const (
	PaymentStripe PaymentMethod = "stripe"
	PaymentPayPal PaymentMethod = "paypal"
)

type CreatePaymentInput struct {
	OrderID   int
	BuyerID   int
	Amount    float64
	Currency  string
	Method    PaymentMethod
	ReturnURL string // PayPal approval return URL
	CancelURL string // PayPal cancel URL
}

type PaymentResult struct {
	PaymentID         int64
	Status            string
	ClientSecret      string // Stripe PaymentIntent client_secret
	PayPalApprovalURL string // PayPal approval URL
	GatewayOrderID    string
}
