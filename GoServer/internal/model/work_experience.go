package model

// WorkExperienceItem 工作经验条目
type WorkExperienceItem struct {
	ID          int    `json:"id"`
	SectionType string `json:"sectionType"`
	ItemKey     string `json:"itemKey"`
	Content     string `json:"content"`
	SortOrder   int    `json:"sortOrder"`
}

// WorkExperienceData 工作经验完整数据(按区块分组)
type WorkExperienceData struct {
	Hero       map[string]string `json:"hero"`
	Features24 map[string]string `json:"features24"`
	Features25 map[string]string `json:"features25"`
	Steps2     map[string]string `json:"steps2"`
	Contact10  map[string]string `json:"contact10"`
}
