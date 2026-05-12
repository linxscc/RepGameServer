package service

import (
	"GoServer/internal/model"
	"fmt"
)

// GetAllWorkExperienceItems 获取所有工作经验条目
func GetAllWorkExperienceItems() ([]model.WorkExperienceItem, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `SELECT id, section_type, item_key, content, sort_order
	          FROM work_experience_items ORDER BY section_type, sort_order`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query work_experience_items: %v", err)
	}
	defer rows.Close()

	var items []model.WorkExperienceItem
	for rows.Next() {
		var item model.WorkExperienceItem
		if err := rows.Scan(&item.ID, &item.SectionType, &item.ItemKey,
			&item.Content, &item.SortOrder); err != nil {
			return nil, fmt.Errorf("scan work_experience_items: %v", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// BuildWorkExperienceData 将扁平条目组装为按区块分组的Map
func BuildWorkExperienceData(items []model.WorkExperienceItem) *model.WorkExperienceData {
	data := &model.WorkExperienceData{
		Hero:       make(map[string]string),
		Features24: make(map[string]string),
		Features25: make(map[string]string),
		Steps2:     make(map[string]string),
		Contact10:  make(map[string]string),
	}
	for _, item := range items {
		switch item.SectionType {
		case "hero":
			data.Hero[item.ItemKey] = item.Content
		case "features24":
			data.Features24[item.ItemKey] = item.Content
		case "features25":
			data.Features25[item.ItemKey] = item.Content
		case "steps2":
			data.Steps2[item.ItemKey] = item.Content
		case "contact10":
			data.Contact10[item.ItemKey] = item.Content
		}
	}
	return data
}
