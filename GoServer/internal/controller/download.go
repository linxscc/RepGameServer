package controller

import (
	v1 "GoServer/api/v1"
	"GoServer/internal/service"
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

type Download struct{}

// GetDownloadItems GET /download-items
func (d *Download) GetDownloadItems(ctx context.Context, req *v1.GetDownloadItemsReq) (res *v1.GetDownloadItemsRes, err error) {
	items, err := service.GetAllDownloadItems()
	if err != nil {
		g.Log().Errorf(ctx, "GetDownloadItems error: %v", err)
		return nil, err
	}

	data := make(v1.GetDownloadItemsRes, 0, len(items))
	for _, item := range items {
		data = append(data, v1.DownloadItemResponse{
			ID:          item.ID,
			Name:        item.Name,
			Version:     item.Version,
			Size:        item.SizeMb,
			Description: item.Description,
			DownloadURL: item.DownloadURL,
			Icon:        item.Icon,
			OsType:      item.OsType,
		})
	}
	return &data, nil
}

// GetSystemRequirements GET /system-requirements
func (d *Download) GetSystemRequirements(ctx context.Context, req *v1.GetSystemRequirementsReq) (res *v1.GetSystemRequirementsRes, err error) {
	items, err := service.GetAllSystemRequirements()
	if err != nil {
		g.Log().Errorf(ctx, "GetSystemRequirements error: %v", err)
		return nil, err
	}

	data := make(v1.GetSystemRequirementsRes, 0, len(items))
	for _, item := range items {
		data = append(data, v1.SystemRequirementResponse{
			ID:           item.ID,
			OsType:       item.OsType,
			OsLabel:      item.OsLabel,
			Requirements: service.ParseRequirements(item.Requirements),
		})
	}
	return &data, nil
}
