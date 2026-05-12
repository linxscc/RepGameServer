package service

import (
	"GoServer/internal/model"
	"encoding/json"
	"fmt"
)

// GetAllDownloadItems 获取所有启用的下载项
func GetAllDownloadItems() ([]model.DownloadItem, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `SELECT id, name, version, size_mb, COALESCE(description,''),
	          download_url, icon, os_type, sort_order, is_active
	          FROM download_items WHERE is_active = 1 ORDER BY sort_order`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query download_items: %v", err)
	}
	defer rows.Close()

	var items []model.DownloadItem
	for rows.Next() {
		var item model.DownloadItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Version, &item.SizeMb,
			&item.Description, &item.DownloadURL, &item.Icon,
			&item.OsType, &item.SortOrder, &item.IsActive); err != nil {
			return nil, fmt.Errorf("scan download_items: %v", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// GetAllSystemRequirements 获取所有系统要求
func GetAllSystemRequirements() ([]model.SystemRequirement, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `SELECT id, os_type, os_label, requirements, sort_order
	          FROM system_requirements ORDER BY sort_order`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query system_requirements: %v", err)
	}
	defer rows.Close()

	var items []model.SystemRequirement
	for rows.Next() {
		var item model.SystemRequirement
		if err := rows.Scan(&item.ID, &item.OsType, &item.OsLabel,
			&item.Requirements, &item.SortOrder); err != nil {
			return nil, fmt.Errorf("scan system_requirements: %v", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ParseRequirements 将JSON字符串解析为[]string
func ParseRequirements(raw string) []string {
	var list []string
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		return []string{}
	}
	return list
}
