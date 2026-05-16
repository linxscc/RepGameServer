package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type paypalClient struct {
	clientID  string
	secret    string
	apiBase   string
	token     string
	tokenExp  time.Time
	http      *http.Client
}

var ppClient *paypalClient

func InitPayPal() {
	cid := os.Getenv("PAYPAL_CLIENT_ID")
	sec := os.Getenv("PAYPAL_SECRET_KEY")
	if cid == "" || sec == "" {
		return
	}
	mode := os.Getenv("PAYPAL_MODE")
	appEnv := os.Getenv("APP_ENV")
	base := "https://api-m.paypal.com"
	if mode == "sandbox" || (mode == "" && appEnv != "production") {
		base = "https://api-m.sandbox.paypal.com"
	}
	ppClient = &paypalClient{
		clientID: cid,
		secret:   sec,
		apiBase:  base,
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *paypalClient) ensureToken() error {
	if c.token != "" && time.Now().Before(c.tokenExp) {
		return nil
	}
	req, _ := http.NewRequest("POST", c.apiBase+"/v1/oauth2/token", bytes.NewReader([]byte("grant_type=client_credentials")))
	req.SetBasicAuth(c.clientID, c.secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("paypal token request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("paypal token error status %d: %s", resp.StatusCode, string(body))
	}
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("paypal token parse: %v", err)
	}
	c.token = result.AccessToken
	c.tokenExp = time.Now().Add(time.Duration(result.ExpiresIn-60) * time.Second)
	return nil
}

func (c *paypalClient) post(path string, payload, result interface{}) error {
	if err := c.ensureToken(); err != nil {
		return err
	}
	var body []byte
	if payload != nil {
		body, _ = json.Marshal(payload)
	}
	req, _ := http.NewRequest("POST", c.apiBase+path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("paypal request %s: %v", path, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("paypal %s status %d: %s", path, resp.StatusCode, string(respBody))
	}
	if result != nil {
		return json.Unmarshal(respBody, result)
	}
	return nil
}

// ── Public API ──

func VerifyPayPalWebhookSignature(headers map[string]string, body []byte) error {
	if ppClient == nil {
		return fmt.Errorf("PayPal client not initialized")
	}

	webhookID := os.Getenv("PAYPAL_WEBHOOK_ID")
	if webhookID == "" {
		return fmt.Errorf("PAYPAL_WEBHOOK_ID not set")
	}

	var event json.RawMessage
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("parse webhook body: %v", err)
	}

	req := map[string]interface{}{
		"auth_algo":         headers["Paypal-Auth-Algo"],
		"cert_url":          headers["Paypal-Cert-Url"],
		"transmission_id":   headers["Paypal-Transmission-Id"],
		"transmission_sig":  headers["Paypal-Transmission-Sig"],
		"transmission_time": headers["Paypal-Transmission-Time"],
		"webhook_id":        webhookID,
		"webhook_event":     event,
	}

	var result struct {
		VerificationStatus string `json:"verification_status"`
	}
	if err := ppClient.post("/v1/notifications/verify-webhook-signature", req, &result); err != nil {
		return fmt.Errorf("paypal verify webhook: %v", err)
	}
	if result.VerificationStatus != "SUCCESS" {
		return fmt.Errorf("paypal webhook verification failed: %s", result.VerificationStatus)
	}
	return nil
}

func CreatePayPalPayment(input CreatePaymentInput) (*PaymentResult, error) {
	if ppClient == nil {
		return nil, fmt.Errorf("PayPal client not initialized (missing API keys)")
	}

	type paypalAmt struct {
		Currency string `json:"currency_code"`
		Value    string `json:"value"`
	}
	type purchaseUnit struct {
		ReferenceID string      `json:"reference_id"`
		Amount      *paypalAmt  `json:"amount"`
		Description string      `json:"description"`
	}
	type appContext struct {
		ReturnURL string `json:"return_url"`
		CancelURL string `json:"cancel_url"`
	}
	type createOrderReq struct {
		Intent             string          `json:"intent"`
		PurchaseUnits      []purchaseUnit  `json:"purchase_units"`
		ApplicationContext *appContext     `json:"application_context,omitempty"`
	}

	req := createOrderReq{
		Intent: "CAPTURE",
		PurchaseUnits: []purchaseUnit{
			{
				ReferenceID: fmt.Sprintf("order_%d", input.OrderID),
				Amount: &paypalAmt{
					Currency: input.Currency,
					Value:    fmt.Sprintf("%.2f", float64(input.Amount)/100),
				},
				Description: fmt.Sprintf("Order #%d", input.OrderID),
			},
		},
		ApplicationContext: &appContext{
			ReturnURL: input.ReturnURL,
			CancelURL: input.CancelURL,
		},
	}

	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Links  []struct {
			Rel  string `json:"rel"`
			Href string `json:"href"`
		} `json:"links"`
	}
	if err := ppClient.post("/v2/checkout/orders", req, &result); err != nil {
		return nil, err
	}

	var approvalURL string
	for _, link := range result.Links {
		if link.Rel == "approve" {
			approvalURL = link.Href
			break
		}
	}

	return &PaymentResult{
		Status:            result.Status,
		PayPalApprovalURL: approvalURL,
		GatewayOrderID:    result.ID,
	}, nil
}

func CapturePayPalOrder(paypalOrderID string) (int64, error) {
	if ppClient == nil {
		return 0, fmt.Errorf("PayPal client not initialized")
	}

	var result struct {
		Status        string `json:"status"`
		PurchaseUnits []struct {
			Payments struct {
				Captures []struct {
					Amount struct {
						CurrencyCode string `json:"currency_code"`
						Value        string `json:"value"`
					} `json:"amount"`
				} `json:"captures"`
			} `json:"payments"`
		} `json:"purchase_units"`
	}
	if err := ppClient.post("/v2/checkout/orders/"+paypalOrderID+"/capture", nil, &result); err != nil {
		return 0, err
	}
	if result.Status != "COMPLETED" {
		return 0, fmt.Errorf("paypal capture status: %s", result.Status)
	}
	if len(result.PurchaseUnits) == 0 || len(result.PurchaseUnits[0].Payments.Captures) == 0 {
		return 0, fmt.Errorf("no captures in PayPal response")
	}

	capturedDollars := result.PurchaseUnits[0].Payments.Captures[0].Amount.Value
	var dollars float64
	if _, err := fmt.Sscanf(capturedDollars, "%f", &dollars); err != nil {
		return 0, fmt.Errorf("parse captured amount: %v", err)
	}
	return DollarsToCents(dollars), nil
}

func ProcessPayPalWebhook(paypalOrderID string) error {
	store := NewIdempotencyStore()
	key := fmt.Sprintf("paypal:%s", paypalOrderID)

	exists, err := store.CheckAndSet(key)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	if _, err := CapturePayPalOrder(paypalOrderID); err != nil {
		return err
	}

	db, err := GetDB()
	if err != nil {
		return err
	}
	
	var orderID int
	err = db.QueryRow(`
		SELECT o.id
		FROM voyara_orders o
		JOIN voyara_payments p ON p.order_id = o.id
		WHERE p.paypal_order_id = ?`, paypalOrderID).
		Scan(&orderID)
	if err != nil {
		return fmt.Errorf("find order by paypal order: %v", err)
	}

	_, _ = db.Exec(`UPDATE voyara_orders SET payment_status = 'paid', paid_at = NOW() WHERE id = ?`, orderID)
	_, _ = db.Exec(`UPDATE voyara_payments SET payment_status = 'succeeded', paid_at = NOW() WHERE order_id = ?`, orderID)

	// Async payment success email (failure does not block payment)
	var buyerEmail, orderNo string
	_ = db.QueryRow(`SELECT u.email, o.order_no FROM voyara_users u JOIN voyara_orders o ON o.buyer_id = u.id WHERE o.id = ?`, orderID).Scan(&buyerEmail, &orderNo)
	if buyerEmail != "" {
		SendPaymentSuccessEmail(buyerEmail, orderNo)
	}

	return nil
}
