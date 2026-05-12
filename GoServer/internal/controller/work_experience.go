package controller

import (
	v1 "GoServer/api/v1"
	"GoServer/internal/service"
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

type WorkExperience struct{}

// GetWorkExperience GET /work-experience
func (w *WorkExperience) GetWorkExperience(ctx context.Context, req *v1.GetWorkExperienceReq) (res *v1.GetWorkExperienceRes, err error) {
	items, err := service.GetAllWorkExperienceItems()
	if err != nil {
		g.Log().Errorf(ctx, "GetWorkExperience error: %v", err)
		return nil, err
	}

	data := service.BuildWorkExperienceData(items)

	return &v1.WorkExperienceDataResponse{
		Hero:       data.Hero,
		Features24: data.Features24,
		Features25: data.Features25,
		Steps2:     data.Steps2,
		Contact10:  data.Contact10,
	}, nil
}
