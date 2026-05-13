package service

import (
	"context"
	"fmt"
	"os"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/webhook"
)

func init() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}

func CreateStripePayment(input CreatePaymentInput) (*PaymentResult, error) {
	params := &stripe.PaymentIntentParams{
		Params: stripe.Params{
			Metadata: map[string]string{
				"order_id": fmt.Sprintf("%d", input.OrderID),
			},
		},
		Amount:   stripe.Int64(int64(input.Amount * 100)),
		Currency: stripe.String(string(input.Currency)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe create payment intent: %v", err)
	}

	return &PaymentResult{
		Status:         string(pi.Status),
		ClientSecret:   pi.ClientSecret,
		GatewayOrderID: pi.ID,
	}, nil
}

func VerifyStripeWebhook(payload []byte, sigHeader string) (string, int, error) {
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return "dev_skip", 0, nil
	}

	event, err := webhook.ConstructEvent(payload, sigHeader, webhookSecret)
	if err != nil {
		return "", 0, fmt.Errorf("stripe webhook verification failed: %v", err)
	}

	var orderID int
	switch event.Type {
	case "payment_intent.succeeded":
	case "payment_intent.payment_failed":
	}

	return event.Type, orderID, nil
}

func ProcessStripeWebhookEvent(eventType string, orderID int) error {
	ctx := context.Background()
	store := NewIdempotencyStore()
	key := fmt.Sprintf("stripe:%s:%d", eventType, orderID)

	exists, err := store.CheckAndSet(key)
	if err != nil {
		return err
	}
	if exists {
		g.Log().Infof(ctx, "Stripe webhook already processed: %s", key)
		return nil
	}

	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	switch eventType {
	case "payment_intent.succeeded":
		_, err = db.Exec(`UPDATE voyara_orders SET payment_status = 'paid', paid_at = NOW() WHERE id = ? AND payment_status = 'pending'`, orderID)
		if err != nil {
			return fmt.Errorf("update order payment: %v", err)
		}
		_, err = db.Exec(`INSERT INTO voyara_payments (order_id, buyer_id, amount, payment_method, payment_status, paid_at)
			SELECT id, buyer_id, grand_total, 'stripe', 'succeeded', NOW() FROM voyara_orders WHERE id = ?`, orderID)
		if err != nil {
			return fmt.Errorf("insert payment: %v", err)
		}
	case "payment_intent.payment_failed":
		_, err = db.Exec(`UPDATE voyara_orders SET payment_status = 'pending' WHERE id = ?`, orderID)
		if err != nil {
			return fmt.Errorf("update order payment failed: %v", err)
		}
	}

	return nil
}
