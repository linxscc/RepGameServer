package model

// DownloadItem 下载项
type DownloadItem struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	SizeMb      string `json:"size"`
	Description string `json:"description"`
	DownloadURL string `json:"downloadUrl"`
	Icon        string `json:"icon"`
	OsType      string `json:"osType"`
	SortOrder   int    `json:"sortOrder"`
	IsActive    int    `json:"isActive"`
}

// SystemRequirement 系统要求
type SystemRequirement struct {
	ID           int    `json:"id"`
	OsType       string `json:"osType"`
	OsLabel      string `json:"osLabel"`
	Requirements string `json:"requirements"` // JSON array string
	SortOrder    int    `json:"sortOrder"`
}
