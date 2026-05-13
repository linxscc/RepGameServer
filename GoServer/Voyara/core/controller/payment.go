package controller

import (
	"context"
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
	defer db.Close()

	var amount float64
	err = db.QueryRow(`SELECT grand_total FROM voyara_orders WHERE id = ? AND buyer_id = ? AND payment_status = 'pending'`, req.OrderID, userID).Scan(&amount)
	if err != nil {
		return nil, fmt.Errorf("order not found or already paid")
	}

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
	defer db.Close()

	var amount float64
	err = db.QueryRow(`SELECT grand_total FROM voyara_orders WHERE id = ? AND buyer_id = ? AND payment_status = 'pending'`, req.OrderID, userID).Scan(&amount)
	if err != nil {
		return nil, fmt.Errorf("order not found or already paid")
	}

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
	if err := service.CapturePayPalOrder(req.PayPalOrderID); err != nil {
		return nil, err
	}

	db, err := service.GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	_, _ = db.Exec(`UPDATE voyara_orders SET payment_status = 'paid', paid_at = NOW() WHERE id = ? AND payment_status = 'pending'`, req.OrderID)
	_, _ = db.Exec(`UPDATE voyara_payments SET payment_status = 'succeeded', paid_at = NOW() WHERE paypal_order_id = ?`, req.PayPalOrderID)

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

	if event.EventType == "CHECKOUT.ORDER.APPROVED" {
		if err := service.ProcessPayPalWebhook(event.Resource.ID); err != nil {
			g.Log().Errorf(r.Context(), "PayPal webhook: process error: %v", err)
			r.Response.WriteStatus(500)
			return
		}
	}

	r.Response.WriteStatus(200)
}
