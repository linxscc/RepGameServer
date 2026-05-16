package service

import "fmt"

type PaymentMethod string

const (
	PaymentStripe PaymentMethod = "stripe"
	PaymentPayPal PaymentMethod = "paypal"
)

type CreatePaymentInput struct {
	OrderID   int
	BuyerID   int
	Amount    int64
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

func RecordPayment(orderID, buyerID int, amount int64, method, gatewayID string) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	
	col := "stripe_payment_intent_id"
	if method == "paypal" {
		col = "paypal_order_id"
	}

	_, err = db.Exec(fmt.Sprintf(`
		INSERT INTO voyara_payments (order_id, buyer_id, amount, currency, payment_method, payment_status, %s)
		VALUES (?, ?, ?, 'USD', ?, 'pending', ?)`, col),
		orderID, buyerID, CentsToDollars(amount), method, gatewayID)
	return err
}
