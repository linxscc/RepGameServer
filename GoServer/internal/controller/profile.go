package controller

import (
	v1 "GoServer/api/v1"
	"GoServer/internal/service"
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

type Profile struct{}

// GetProfileInfo GET /profile-info
func (p *Profile) GetProfileInfo(ctx context.Context, req *v1.GetProfileInfoReq) (res *v1.GetProfileInfoRes, err error) {
	info, err := service.GetProfileInfo()
	if err != nil {
		g.Log().Errorf(ctx, "GetProfileInfo error: %v", err)
		return nil, err
	}

	return &v1.ProfileInfoResponse{
		FullName:  info.FullName,
		Title:     info.Title,
		Tagline:   info.Tagline,
		AboutText: info.AboutText,
		Email:     info.Email,
		Phone:     info.Phone,
		Languages: info.Languages,
	}, nil
}
