package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"

	v1 "GoServer/Voyara/api/v1"
	"GoServer/Voyara/core/service"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type Payment struct{}

func (c *Payment) CreatePayment(ctx context.Context, req *v1.CreatePaymentReq) (res *v1.CreatePaymentRes, err error) {
	userID := ctx.Value("userID").(int)

	if service.PaymentMethod(req.Method) == service.PaymentPayPal {
		return c.createPayPalPayment(ctx, userID, req)
	}
	return c.createStripePayment(ctx, userID, req)
}

func (c *Payment) createStripePayment(ctx context.Context, userID int, req *v1.CreatePaymentReq) (*v1.CreatePaymentRes, error) {
	db, err := service.GetDB()
	if err != nil {
		return nil, err
	}
	
	var amountF64 float64
	var paymentStatus string
	var snapshotItems sql.NullString
	err = db.QueryRow(`SELECT grand_total, payment_status, snapshot_items FROM voyara_orders WHERE id = ? AND buyer_id = ?`, req.OrderID, userID).Scan(&amountF64, &paymentStatus, &snapshotItems)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found")
	}
	if err != nil {
		return nil, err
	}
	if paymentStatus != "pending" {
		return nil, fmt.Errorf("order payment status is '%s', expected 'pending'", paymentStatus)
	}
	if !snapshotItems.Valid || snapshotItems.String == "" || snapshotItems.String == "[]" {
		return nil, fmt.Errorf("order snapshot missing")
	}
	amount := service.DollarsToCents(amountF64)

	result, err := service.CreateStripePayment(service.CreatePaymentInput{
		OrderID:  req.OrderID,
		BuyerID:  userID,
		Amount:   amount,
		Currency: "USD",
		Method:   service.PaymentStripe,
	})
	if err != nil {
		return nil, err
	}

	_ = service.RecordPayment(req.OrderID, userID, amount, "stripe", result.GatewayOrderID)

	return &v1.CreatePaymentRes{
		PaymentID:      0,
		ClientSecret:   result.ClientSecret,
		GatewayOrderID: result.GatewayOrderID,
		Status:         result.Status,
	}, nil
}

func (c *Payment) createPayPalPayment(ctx context.Context, userID int, req *v1.CreatePaymentReq) (*v1.CreatePaymentRes, error) {
	db, err := service.GetDB()
	if err != nil {
		return nil, err
	}
	
	var amountF64 float64
	var paymentStatus string
	var snapshotItems sql.NullString
	err = db.QueryRow(`SELECT grand_total, payment_status, snapshot_items FROM voyara_orders WHERE id = ? AND buyer_id = ?`, req.OrderID, userID).Scan(&amountF64, &paymentStatus, &snapshotItems)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found")
	}
	if err != nil {
		return nil, err
	}
	if paymentStatus != "pending" {
		return nil, fmt.Errorf("order payment status is '%s', expected 'pending'", paymentStatus)
	}
	if !snapshotItems.Valid || snapshotItems.String == "" || snapshotItems.String == "[]" {
		return nil, fmt.Errorf("order snapshot missing")
	}
	amount := service.DollarsToCents(amountF64)

	result, err := service.CreatePayPalPayment(service.CreatePaymentInput{
		OrderID:   req.OrderID,
		BuyerID:   userID,
		Amount:    amount,
		Currency:  "USD",
		Method:    service.PaymentPayPal,
		ReturnURL: req.ReturnURL,
		CancelURL: req.CancelURL,
	})
	if err != nil {
		return nil, err
	}

	_ = service.RecordPayment(req.OrderID, userID, amount, "paypal", result.GatewayOrderID)

	return &v1.CreatePaymentRes{
		PaymentID:         0,
		PayPalApprovalURL: result.PayPalApprovalURL,
		GatewayOrderID:    result.GatewayOrderID,
		Status:            result.Status,
	}, nil
}

func (c *Payment) CapturePayPal(ctx context.Context, req *v1.CapturePayPalReq) (res *v1.MessageRes, err error) {
	userID := ctx.Value("userID").(int)

	store := service.NewIdempotencyStore()
	key := fmt.Sprintf("capture:paypal:%s", req.PayPalOrderID)
	exists, err := store.CheckAndSet(key)
	if err != nil {
		return nil, err
	}
	if exists {
		return &v1.MessageRes{Message: "Payment already captured"}, nil
	}

	db, err := service.GetDB()
	if err != nil {
		return nil, err
	}
	
	var orderID int
	var buyerID int
	var paymentStatus string
	var grandTotalF64 float64
	err = db.QueryRow(`
		SELECT o.id, o.buyer_id, o.payment_status, o.grand_total
		FROM voyara_orders o
		JOIN voyara_payments p ON p.order_id = o.id
		WHERE p.paypal_order_id = ?`, req.PayPalOrderID).
		Scan(&orderID, &buyerID, &paymentStatus, &grandTotalF64)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found for this PayPal payment")
	}
	if err != nil {
		return nil, err
	}
	if buyerID != userID {
		return nil, fmt.Errorf("unauthorized: order does not belong to current user")
	}
	if paymentStatus != "pending" {
		return nil, fmt.Errorf("order payment status is '%s', expected 'pending'", paymentStatus)
	}

	capturedAmount, err := service.CapturePayPalOrder(req.PayPalOrderID)
	if err != nil {
		return nil, err
	}

	localAmount := service.DollarsToCents(grandTotalF64)
	if capturedAmount != localAmount {
		return nil, fmt.Errorf("paypal amount mismatch: captured %d, expected %d", capturedAmount, localAmount)
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE voyara_orders SET payment_status = 'paid', paid_at = NOW() WHERE id = ?`, orderID)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(`UPDATE voyara_payments SET payment_status = 'succeeded', paid_at = NOW() WHERE paypal_order_id = ?`, req.PayPalOrderID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &v1.MessageRes{Message: "Payment captured successfully"}, nil
}

func (c *Payment) StripeWebhook(r *ghttp.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		g.Log().Errorf(r.Context(), "Stripe webhook: read body error: %v", err)
		r.Response.WriteStatus(400)
		return
	}

	sigHeader := r.Header.Get("Stripe-Signature")
	eventType, orderID, err := service.VerifyStripeWebhook(payload, sigHeader)
	if err != nil {
		g.Log().Errorf(r.Context(), "Stripe webhook: verification error: %v", err)
		r.Response.WriteStatus(400)
		return
	}

	if err := service.ProcessStripeWebhookEvent(eventType, orderID); err != nil {
		g.Log().Errorf(r.Context(), "Stripe webhook: process error: %v", err)
		r.Response.WriteStatus(500)
		return
	}

	r.Response.WriteStatus(200)
}

func (c *Payment) PayPalWebhook(r *ghttp.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		g.Log().Errorf(r.Context(), "PayPal webhook: read body error: %v", err)
		r.Response.WriteStatus(400)
		return
	}

	headers := map[string]string{
		"Paypal-Transmission-Id":   r.Header.Get("Paypal-Transmission-Id"),
		"Paypal-Transmission-Time": r.Header.Get("Paypal-Transmission-Time"),
		"Paypal-Cert-Url":          r.Header.Get("Paypal-Cert-Url"),
		"Paypal-Auth-Algo":         r.Header.Get("Paypal-Auth-Algo"),
		"Paypal-Transmission-Sig":  r.Header.Get("Paypal-Transmission-Sig"),
	}

	if err := service.VerifyPayPalWebhookSignature(headers, payload); err != nil {
		g.Log().Errorf(r.Context(), "PayPal webhook: verification error: %v", err)
		r.Response.WriteStatus(401)
		return
	}

	var event struct {
		EventType string `json:"event_type"`
		Resource  struct {
			ID string `json:"id"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		g.Log().Errorf(r.Context(), "PayPal webhook: parse error: %v", err)
		r.Response.WriteStatus(400)
		return
	}

	switch event.EventType {
	case "CHECKOUT.ORDER.APPROVED":
		if err := service.ProcessPayPalWebhook(event.Resource.ID); err != nil {
			g.Log().Errorf(r.Context(), "PayPal webhook: process error: %v", err)
			r.Response.WriteStatus(500)
			return
		}
	default:
		g.Log().Infof(r.Context(), "PayPal webhook: unhandled event type: %s", event.EventType)
	}

	r.Response.WriteStatus(200)
}
