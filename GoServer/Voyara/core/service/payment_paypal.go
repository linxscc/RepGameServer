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

func initPayPal() {
	cid := os.Getenv("PAYPAL_CLIENT_ID")
	sec := os.Getenv("PAYPAL_SECRET_KEY")
	if cid == "" || sec == "" {
		return
	}
	base := "https://api-m.paypal.com"
	if os.Getenv("APP_ENV") != "production" {
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
					Value:    fmt.Sprintf("%.2f", input.Amount),
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

func CapturePayPalOrder(paypalOrderID string) error {
	if ppClient == nil {
		return fmt.Errorf("PayPal client not initialized")
	}

	var result struct {
		Status string `json:"status"`
	}
	if err := ppClient.post("/v2/checkout/orders/"+paypalOrderID+"/capture", nil, &result); err != nil {
		return err
	}
	if result.Status != "COMPLETED" {
		return fmt.Errorf("paypal capture status: %s", result.Status)
	}
	return nil
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

	if err := CapturePayPalOrder(paypalOrderID); err != nil {
		return err
	}

	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

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

	return nil
}
