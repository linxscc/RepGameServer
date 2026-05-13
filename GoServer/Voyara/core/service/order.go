package service

import (
	"GoServer/Voyara/core/model"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func generateOrderNo() string {
	return fmt.Sprintf("V%s%06d", time.Now().Format("060102"), time.Now().UnixNano()%1000000)
}

func Checkout(buyerID int, productIDs []int, addr ShippingAddr, idempotencyKey string) (*model.Order, error) {
	if idempotencyKey != "" {
		store := NewIdempotencyStore()
		exists, err := store.CheckAndSet(idempotencyKey)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("duplicate request")
		}
	}

	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	placeholders := make([]string, len(productIDs))
	args := make([]interface{}, 0, len(productIDs)+1)
	args = append(args, buyerID)
	for i, pid := range productIDs {
		placeholders[i] = "?"
		args = append(args, pid)
	}

	rows, err := db.Query(`
		SELECT c.product_id, c.quantity, p.title, p.price, p.seller_id, COALESCE(p.images,'[]'),
		       COALESCE(s.shop_name,'')
		FROM voyara_cart_items c
		JOIN voyara_products p ON c.product_id = p.id
		LEFT JOIN voyara_sellers s ON p.seller_id = s.id
		WHERE c.user_id = ? AND c.selected = 1 AND c.product_id IN (`+strings.Join(placeholders, ",")+`)
		AND p.status = 'active'`, args...)
	if err != nil {
		return nil, fmt.Errorf("query cart items: %v", err)
	}
	defer rows.Close()

	type checkoutItem struct {
		productID int
		quantity  int
		title     string
		price     float64
		sellerID  int
		imageURL  string
	}
	var items []checkoutItem
	sellerIDs := make(map[int]bool)
	for rows.Next() {
		var item checkoutItem
		var imagesStr, shopName string
		if err := rows.Scan(&item.productID, &item.quantity, &item.title, &item.price, &item.sellerID, &imagesStr, &shopName); err != nil {
			return nil, fmt.Errorf("scan cart item: %v", err)
		}
		if imagesStr != "" && imagesStr != "[]" {
			var imgs []string
			if json.Unmarshal([]byte(imagesStr), &imgs) == nil && len(imgs) > 0 {
				item.imageURL = imgs[0]
			}
		}
		items = append(items, item)
		sellerIDs[item.sellerID] = true
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no available items selected")
	}
	if len(sellerIDs) > 1 {
		return nil, fmt.Errorf("items must be from the same seller")
	}

	var sellerID int
	for sid := range sellerIDs {
		sellerID = sid
	}

	var subtotal float64
	for _, item := range items {
		subtotal += item.price * float64(item.quantity)
	}
	grandTotal := subtotal
	itemCount := len(items)

	addrBytes, _ := json.Marshal(addr)
	orderNo := generateOrderNo()

	snapshot := make([]map[string]interface{}, len(items))
	for i, item := range items {
		snapshot[i] = map[string]interface{}{
			"productId": item.productID,
			"title":     item.title,
			"price":     item.price,
			"quantity":  item.quantity,
			"imageUrl":  item.imageURL,
		}
	}
	snapshotJSON, _ := json.Marshal(snapshot)

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %v", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		INSERT INTO voyara_orders (order_no, buyer_id, seller_id, product_id, item_count, amount, subtotal, grand_total, shipping_address, payment_status, shipping_status, snapshot_items)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending', 'pending', ?)`,
		orderNo, buyerID, sellerID, items[0].productID, itemCount, grandTotal, subtotal, grandTotal, string(addrBytes), string(snapshotJSON))
	if err != nil {
		return nil, fmt.Errorf("insert order: %v", err)
	}
	orderID, _ := res.LastInsertId()

	for _, item := range items {
		total := item.price * float64(item.quantity)
		_, err = tx.Exec(`
			INSERT INTO voyara_order_items (order_id, product_id, title, price, quantity, total, image_url)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			orderID, item.productID, item.title, item.price, item.quantity, total, item.imageURL)
		if err != nil {
			return nil, fmt.Errorf("insert order item: %v", err)
		}
	}

	pidArgs := make([]interface{}, 0, len(productIDs))
	for _, pid := range productIDs {
		pidArgs = append(pidArgs, pid)
	}
	pidArgs = append(pidArgs, buyerID)
	_, err = tx.Exec(`DELETE FROM voyara_cart_items WHERE product_id IN (`+strings.Join(placeholders, ",")+`) AND user_id = ?`, pidArgs...)
	if err != nil {
		return nil, fmt.Errorf("clear cart: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %v", err)
	}

	return &model.Order{
		ID:              int(orderID),
		OrderNo:         orderNo,
		BuyerID:         buyerID,
		SellerID:        sellerID,
		Amount:          grandTotal,
		Subtotal:        subtotal,
		Currency:        "USD",
		PaymentStatus:   "pending",
		ShippingStatus:  "pending",
		ShippingAddress: sql.NullString{String: string(addrBytes), Valid: true},
	}, nil
}

func GetOrdersByBuyer(buyerID int) ([]model.Order, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT o.id, o.order_no, o.buyer_id, o.seller_id, o.item_count,
		       o.amount, o.subtotal, o.currency,
		       o.payment_status, o.shipping_status, COALESCE(o.tracking_number,''),
		       COALESCE(o.shipping_address,''), o.created_at
		FROM voyara_orders o
		WHERE o.buyer_id = ?
		ORDER BY o.created_at DESC`, buyerID)
	if err != nil {
		return nil, fmt.Errorf("query orders: %v", err)
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		var addrStr string
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.BuyerID, &o.SellerID, &o.ItemCount,
			&o.Amount, &o.Subtotal, &o.Currency,
			&o.PaymentStatus, &o.ShippingStatus, &o.TrackingNumber,
			&addrStr, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan order: %v", err)
		}
		o.ShippingAddress = sql.NullString{String: addrStr, Valid: true}
		_ = loadOrderItems(db, &o)
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []model.Order{}
	}
	return orders, nil
}

func loadOrderItems(db *sql.DB, o *model.Order) error {
	rows, err := db.Query(`SELECT id, product_id, title, price, quantity, total, COALESCE(image_url,'') FROM voyara_order_items WHERE order_id = ?`, o.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Title, &item.Price, &item.Quantity, &item.Total, &item.ImageURL); err != nil {
			return err
		}
		o.Items = append(o.Items, item)
	}
	return nil
}

type ShippingAddr struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Country string `json:"country"`
	City    string `json:"city"`
	Street  string `json:"street"`
	ZipCode string `json:"zipCode"`
}

// ShipOrder updates shipping status for an order (seller only)
func ShipOrder(orderID, sellerID int, trackingNumber string) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	res, err := db.Exec(`UPDATE voyara_orders o
		JOIN voyara_products p ON o.product_id = p.id
		SET o.shipping_status = 'shipped', o.tracking_number = ?
		WHERE o.id = ? AND p.seller_id = ? AND o.shipping_status = 'pending'`,
		trackingNumber, orderID, sellerID)
	if err != nil {
		return fmt.Errorf("ship order: %v", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("order not found or already shipped")
	}
	return nil
}

func GetOrdersBySeller(sellerID int) ([]model.Order, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT o.id, o.order_no, o.buyer_id, o.seller_id, o.item_count,
		o.amount, o.subtotal, o.currency,
		o.payment_status, o.shipping_status, COALESCE(o.tracking_number,''),
		COALESCE(o.shipping_address,''), o.created_at
		FROM voyara_orders o
		WHERE o.seller_id = ?
		ORDER BY o.created_at DESC`, sellerID)
	if err != nil {
		return nil, fmt.Errorf("query seller orders: %v", err)
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		var addrStr string
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.BuyerID, &o.SellerID, &o.ItemCount,
			&o.Amount, &o.Subtotal, &o.Currency,
			&o.PaymentStatus, &o.ShippingStatus, &o.TrackingNumber,
			&addrStr, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan order: %v", err)
		}
		o.ShippingAddress = sql.NullString{String: addrStr, Valid: true}
		_ = loadOrderItems(db, &o)
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []model.Order{}
	}
	return orders, nil
}
