package v1

import "github.com/gogf/gf/v2/frame/g"

type CreatePaymentReq struct {
	g.Meta    `path:"/voyara/payments" method:"post" summary:"Create payment" middleware:"auth"`
	OrderID   int    `json:"orderId" v:"required"`
	Method    string `json:"method" v:"required|in:stripe,paypal"`
	ReturnURL string `json:"returnUrl"`
	CancelURL string `json:"cancelUrl"`
}

type CreatePaymentRes struct {
	PaymentID         int64  `json:"paymentId"`
	ClientSecret      string `json:"clientSecret,omitempty"`
	PayPalApprovalURL string `json:"paypalApprovalUrl,omitempty"`
	GatewayOrderID    string `json:"gatewayOrderId"`
	Status            string `json:"status"`
}

type CapturePayPalReq struct {
	g.Meta        `path:"/voyara/payments/paypal/capture" method:"post" summary:"Capture PayPal" middleware:"auth"`
	PayPalOrderID string `json:"paypalOrderId" v:"required"`
	OrderID       int    `json:"orderId" v:"required"`
}

type StripeWebhookReq struct {
	g.Meta `path:"/voyara/payment/stripe-webhook" method:"post" summary:"Stripe webhook"`
}

type PayPalWebhookReq struct {
	g.Meta `path:"/voyara/payment/paypal-webhook" method:"post" summary:"PayPal webhook"`
}
