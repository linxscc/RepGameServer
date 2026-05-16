package service

import (
	"GoServer/Voyara/core/model"
	"database/sql"
	"errors"
	"fmt"
)

func GetCategories() ([]model.Category, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	
	rows, err := db.Query(`SELECT id, name, parent_id, COALESCE(icon,'') FROM voyara_categories ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query categories: %v", err)
	}
	defer rows.Close()

	var cats []model.Category
	for rows.Next() {
		var c model.Category
		var parentID sql.NullInt64
		if err := rows.Scan(&c.ID, &c.Name, &parentID, &c.Icon); err != nil {
			return nil, fmt.Errorf("scan category: %v", err)
		}
		c.ParentID = parentID
		cats = append(cats, c)
	}
	if cats == nil {
		cats = []model.Category{}
	}
	return cats, nil
}

func EnsureSeller(userID int, shopName, description string) (int, error) {
	db, err := GetDB()
	if err != nil {
		return 0, err
	}
	
	var sellerID int
	err = db.QueryRow(`SELECT id FROM voyara_sellers WHERE user_id = ?`, userID).Scan(&sellerID)
	if err == nil {
		return sellerID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("query seller: %v", err)
	}

	res, err := db.Exec(`INSERT INTO voyara_sellers (user_id, shop_name, description) VALUES (?, ?, ?)`,
		userID, shopName, description)
	if err != nil {
		return 0, fmt.Errorf("insert seller: %v", err)
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}
