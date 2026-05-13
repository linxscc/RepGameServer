package service

import (
	"GoServer/Voyara/core/model"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

type CartItemWithProduct struct {
	model.CartItem
	ProductTitle   string  `json:"productTitle"`
	ProductPrice   float64 `json:"productPrice"`
	ProductImage   string  `json:"productImage"`
	ProductStatus  string  `json:"productStatus"`
	SellerID       int     `json:"sellerId"`
	SellerShopName string  `json:"sellerShopName"`
}

func AddToCart(userID, productID, quantity int) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	var status string
	err = db.QueryRow(`SELECT status FROM voyara_products WHERE id = ?`, productID).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("product not found")
	}
	if err != nil {
		return fmt.Errorf("query product: %v", err)
	}
	if status != "active" {
		return fmt.Errorf("product is not available")
	}
	if quantity < 1 {
		quantity = 1
	}
	if quantity > 50 {
		quantity = 50
	}

	_, err = db.Exec(`
		INSERT INTO voyara_cart_items (user_id, product_id, quantity)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE quantity = quantity + ?`, userID, productID, quantity, quantity)
	if err != nil {
		return fmt.Errorf("add to cart: %v", err)
	}
	return nil
}

func UpdateCartItemQuantity(userID, itemID, quantity int) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	if quantity < 1 {
		quantity = 1
	}
	if quantity > 50 {
		quantity = 50
	}

	res, err := db.Exec(`UPDATE voyara_cart_items SET quantity = ? WHERE id = ? AND user_id = ?`, quantity, itemID, userID)
	if err != nil {
		return fmt.Errorf("update cart: %v", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("cart item not found")
	}
	return nil
}

func ToggleCartItemSelected(userID, itemID int, selected bool) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	v := 0
	if selected {
		v = 1
	}
	res, err := db.Exec(`UPDATE voyara_cart_items SET selected = ? WHERE id = ? AND user_id = ?`, v, itemID, userID)
	if err != nil {
		return fmt.Errorf("toggle cart: %v", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("cart item not found")
	}
	return nil
}

func RemoveCartItem(userID, itemID int) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	res, err := db.Exec(`DELETE FROM voyara_cart_items WHERE id = ? AND user_id = ?`, itemID, userID)
	if err != nil {
		return fmt.Errorf("remove cart: %v", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("cart item not found")
	}
	return nil
}

func ClearCart(userID int) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`DELETE FROM voyara_cart_items WHERE user_id = ?`, userID)
	return err
}

func GetCart(userID int) ([]CartItemWithProduct, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT c.id, c.product_id, c.quantity, c.selected,
		       p.title, p.price, p.status,
		       COALESCE(p.images, '[]'),
		       COALESCE(s.id, 0), COALESCE(s.shop_name, '')
		FROM voyara_cart_items c
		JOIN voyara_products p ON c.product_id = p.id
		LEFT JOIN voyara_sellers s ON p.seller_id = s.id
		WHERE c.user_id = ?
		ORDER BY c.created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("query cart: %v", err)
	}
	defer rows.Close()

	var items []CartItemWithProduct
	for rows.Next() {
		var item CartItemWithProduct
		var imagesStr, shopName string
		var sellerID int
		var selected int
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &selected,
			&item.ProductTitle, &item.ProductPrice, &item.ProductStatus,
			&imagesStr, &sellerID, &shopName); err != nil {
			return nil, fmt.Errorf("scan cart: %v", err)
		}
		item.UserID = userID
		item.Selected = selected == 1
		item.SellerID = sellerID
		item.SellerShopName = shopName
		if imagesStr != "" && imagesStr != "[]" {
			var imgs []string
			if json.Unmarshal([]byte(imagesStr), &imgs) == nil && len(imgs) > 0 {
				item.ProductImage = imgs[0]
			}
		}
		items = append(items, item)
	}
	if items == nil {
		items = []CartItemWithProduct{}
	}
	return items, nil
}

func SelectAllCartItems(userID int, selected bool) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	v := 0
	if selected {
		v = 1
	}
	_, err = db.Exec(`UPDATE voyara_cart_items SET selected = ? WHERE user_id = ?`, v, userID)
	return err
}

func GetCartCount(userID int) (int, error) {
	db, err := GetDB()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM voyara_cart_items WHERE user_id = ?`, userID).Scan(&count)
	return count, err
}
