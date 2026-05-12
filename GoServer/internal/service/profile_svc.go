package service

import (
	"GoServer/internal/model"
	"fmt"
)

// GetProfileInfo 获取个人信息（单条记录）
func GetProfileInfo() (*model.ProfileInfo, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `SELECT id, full_name, title, tagline,
	          COALESCE(about_text,''), COALESCE(email,''),
	          COALESCE(phone,''), COALESCE(languages,'')
	          FROM profile_info ORDER BY id LIMIT 1`

	var p model.ProfileInfo
	err = db.QueryRow(query).Scan(&p.ID, &p.FullName, &p.Title, &p.Tagline,
		&p.AboutText, &p.Email, &p.Phone, &p.Languages)
	if err != nil {
		return nil, fmt.Errorf("query profile_info: %v", err)
	}
	return &p, nil
}
