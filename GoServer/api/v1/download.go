package v1

import "github.com/gogf/gf/v2/frame/g"

// ── Download Items ──

type GetDownloadItemsReq struct {
	g.Meta `path:"/download-items" method:"get" summary:"获取下载项列表"`
}

type GetDownloadItemsRes []DownloadItemResponse

type DownloadItemResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Size        string `json:"size"`
	Description string `json:"description"`
	DownloadURL string `json:"downloadUrl"`
	Icon        string `json:"icon"`
	OsType      string `json:"osType"`
}

// ── System Requirements ──

type GetSystemRequirementsReq struct {
	g.Meta `path:"/system-requirements" method:"get" summary:"获取系统要求"`
}

type GetSystemRequirementsRes []SystemRequirementResponse

type SystemRequirementResponse struct {
	ID           int      `json:"id"`
	OsType       string   `json:"osType"`
	OsLabel      string   `json:"osLabel"`
	Requirements []string `json:"requirements"`
}
