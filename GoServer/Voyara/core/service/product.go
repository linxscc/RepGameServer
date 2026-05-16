package service

import (
	"GoServer/Voyara/core/model"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type ProductFilter struct {
	Category  string
	Condition string
	MinPrice  string
	MaxPrice  string
	Search    string
	Page      int
	PageSize  int
}

type PaginatedProducts struct {
	Items    []model.Product
	Total    int
	Page     int
	PageSize int
}

func buildProductWhere(filter ProductFilter) (whereClause string, args []interface{}) {
	var clauses []string
	if filter.Category != "" {
		clauses = append(clauses, "p.category = ?")
		args = append(args, filter.Category)
	}
	if filter.Condition != "" {
		clauses = append(clauses, "p.`condition` = ?")
		args = append(args, filter.Condition)
	}
	if filter.MinPrice != "" {
		clauses = append(clauses, "p.price >= ?")
		args = append(args, filter.MinPrice)
	}
	if filter.MaxPrice != "" {
		clauses = append(clauses, "p.price <= ?")
		args = append(args, filter.MaxPrice)
	}
	if filter.Search != "" {
		clauses = append(clauses, "p.title LIKE ?")
		args = append(args, "%"+filter.Search+"%")
	}
	clauses = append(clauses, "p.status = 'active'")
	whereClause = `WHERE ` + strings.Join(clauses, " AND ")
	return
}

func GetProducts(filter ProductFilter) (*PaginatedProducts, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	whereClause, args := buildProductWhere(filter)

	joinClause := `FROM voyara_products p
		LEFT JOIN voyara_sellers s ON p.seller_id = s.id
		LEFT JOIN voyara_users u ON s.user_id = u.id`

	// Count total matching products
	var total int
	countQuery := `SELECT COUNT(*) ` + joinClause + ` ` + whereClause
	if err := db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count products: %v", err)
	}

	// Apply pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	selectColumns := `SELECT p.id, p.seller_id, p.title, COALESCE(p.description,''), p.price, p.currency,
		p.category, p.` + "`condition`" + `, COALESCE(p.images,'[]'), p.status,
		COALESCE(s.shop_name,''), COALESCE(u.name,'')`

	dataArgs := make([]interface{}, len(args))
	copy(dataArgs, args)
	dataArgs = append(dataArgs, pageSize, offset)

	rows, err := db.Query(selectColumns+` `+joinClause+` `+whereClause+` ORDER BY p.created_at DESC LIMIT ? OFFSET ?`, dataArgs...)
	if err != nil {
		return nil, fmt.Errorf("query products: %v", err)
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		var imagesStr string
		var priceF64 float64
		if err := rows.Scan(&p.ID, &p.SellerID, &p.Title, &p.Description, &priceF64, &p.Currency,
			&p.Category, &p.Condition, &imagesStr, &p.Status, &p.ShopName, &p.SellerName); err != nil {
			return nil, fmt.Errorf("scan product: %v", err)
		}
		p.Price = DollarsToCents(priceF64)
		p.Images = sql.NullString{String: imagesStr, Valid: true}
		products = append(products, p)
	}
	if products == nil {
		products = []model.Product{}
	}
	return &PaginatedProducts{
		Items:    products,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func GetProductByID(id int) (*model.Product, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	var p model.Product
	var imagesStr string
	var priceF64 float64
	err = db.QueryRow(`SELECT p.id, p.seller_id, p.title, COALESCE(p.description,''), p.price, p.currency,
		p.category, p.`+"`condition`"+`, COALESCE(p.images,'[]'), p.status,
		COALESCE(s.shop_name,''), COALESCE(u.name,'')
		FROM voyara_products p
		LEFT JOIN voyara_sellers s ON p.seller_id = s.id
		LEFT JOIN voyara_users u ON s.user_id = u.id
		WHERE p.id = ?`, id).
		Scan(&p.ID, &p.SellerID, &p.Title, &p.Description, &priceF64, &p.Currency,
			&p.Category, &p.Condition, &imagesStr, &p.Status, &p.ShopName, &p.SellerName)
	p.Price = DollarsToCents(priceF64)
	if err != nil {
		return nil, fmt.Errorf("query product: %v", err)
	}
	p.Images = sql.NullString{String: imagesStr, Valid: true}
	return &p, nil
}

func CreateProduct(sellerID int, title, description string, price int64, category, condition string, images []string) (*model.Product, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	imagesJSON := "[]"
	if len(images) > 0 {
		b, _ := json.Marshal(images)
		imagesJSON = string(b)
	}

	res, err := db.Exec(`INSERT INTO voyara_products (seller_id, title, description, price, category, `+"`condition`"+`, images)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, sellerID, title, description, CentsToDollars(price), category, condition, imagesJSON)
	if err != nil {
		return nil, fmt.Errorf("insert product: %v", err)
	}
	id, _ := res.LastInsertId()
	return &model.Product{
		ID:          int(id),
		SellerID:    sellerID,
		Title:       title,
		Description: sql.NullString{String: description, Valid: true},
		Price:       price,
		Currency:    "USD",
		Category:    category,
		Condition:   condition,
		Images:      sql.NullString{String: imagesJSON, Valid: true},
		Status:      "active",
	}, nil
}

func UpdateProduct(id, sellerID int, title, description *string, price *int64, category, condition *string) (*model.Product, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	var sets []string
	var args []interface{}

	if title != nil {
		sets = append(sets, "title = ?")
		args = append(args, *title)
	}
	if description != nil {
		sets = append(sets, "description = ?")
		args = append(args, *description)
	}
	if price != nil {
		sets = append(sets, "price = ?")
		// price is in cents, convert to dollars for DECIMAL column
		args = append(args, CentsToDollars(*price))
	}
	if category != nil {
		sets = append(sets, "category = ?")
		args = append(args, *category)
	}
	if condition != nil {
		sets = append(sets, "`condition` = ?")
		args = append(args, *condition)
	}
	if len(sets) == 0 {
		return GetProductByID(id)
	}

	args = append(args, id, sellerID)
	_, err = db.Exec(`UPDATE voyara_products SET `+strings.Join(sets, ", ")+` WHERE id = ? AND seller_id = ?`, args...)
	if err != nil {
		return nil, fmt.Errorf("update product: %v", err)
	}
	return GetProductByID(id)
}

func GetSellerIDByUserID(userID int) (int, error) {
	db, err := GetDB()
	if err != nil {
		return 0, err
	}

	var sellerID int
	err = db.QueryRow(`SELECT id FROM voyara_sellers WHERE user_id = ?`, userID).Scan(&sellerID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return sellerID, err
}

func GetAllProducts(status string) ([]model.Product, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	query := `SELECT p.id, p.seller_id, p.title, COALESCE(p.description,''), p.price, p.currency,
		p.category, p.` + "`condition`" + `, COALESCE(p.images,'[]'), p.status,
		COALESCE(s.shop_name,''), COALESCE(u.name,''),
		COALESCE(p.created_at,''), COALESCE(p.updated_at,'')
		FROM voyara_products p
		LEFT JOIN voyara_sellers s ON p.seller_id = s.id
		LEFT JOIN voyara_users u ON s.user_id = u.id`

	var args []interface{}
	if status != "" {
		query += ` WHERE p.status = ?`
		args = append(args, status)
	}
	query += ` ORDER BY p.created_at DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query all products: %v", err)
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		var imagesStr string
		var priceF64 float64
		if err := rows.Scan(&p.ID, &p.SellerID, &p.Title, &p.Description, &priceF64, &p.Currency,
			&p.Category, &p.Condition, &imagesStr, &p.Status, &p.ShopName, &p.SellerName,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan product: %v", err)
		}
		p.Images = sql.NullString{String: imagesStr, Valid: true}
		p.Price = DollarsToCents(priceF64)
		products = append(products, p)
	}
	if products == nil {
		products = []model.Product{}
	}
	return products, nil
}

func UpdateProductStatus(productID int, status string) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	res, err := db.Exec(`UPDATE voyara_products SET status = ? WHERE id = ?`, status, productID)
	if err != nil {
		return fmt.Errorf("update product status: %v", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

func GetProductsBySeller(sellerID int) ([]model.Product, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`SELECT p.id, p.seller_id, p.title, COALESCE(p.description,''), p.price, p.currency,
		p.category, p.`+"`condition`"+`, COALESCE(p.images,'[]'), p.status,
		COALESCE(s.shop_name,''), COALESCE(u.name,'')
		FROM voyara_products p
		LEFT JOIN voyara_sellers s ON p.seller_id = s.id
		LEFT JOIN voyara_users u ON s.user_id = u.id
		WHERE p.seller_id = ?
		ORDER BY p.created_at DESC`, sellerID)
	if err != nil {
		return nil, fmt.Errorf("query seller products: %v", err)
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		var imagesStr string
		var priceF64 float64
		if err := rows.Scan(&p.ID, &p.SellerID, &p.Title, &p.Description, &priceF64, &p.Currency,
			&p.Category, &p.Condition, &imagesStr, &p.Status, &p.ShopName, &p.SellerName); err != nil {
			return nil, fmt.Errorf("scan product: %v", err)
		}
		p.Images = sql.NullString{String: imagesStr, Valid: true}
		p.Price = DollarsToCents(priceF64)
		products = append(products, p)
	}
	if products == nil {
		products = []model.Product{}
	}
	return products, nil
}
