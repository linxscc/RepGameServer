package controller

import (
	v1 "GoServer/Voyara/api/v1"
	"GoServer/Voyara/core/service"
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

type Category struct{}

func (c *Category) GetCategories(ctx context.Context, req *v1.GetCategoriesReq) (res *v1.GetCategoriesRes, err error) {
	cats, err := service.GetCategories()
	if err != nil {
		g.Log().Errorf(ctx, "GetCategories error: %v", err)
		return nil, err
	}

	items := make(v1.GetCategoriesRes, 0, len(cats))
	for _, cat := range cats {
		var parentID *int
		if cat.ParentID.Valid {
			pid := int(cat.ParentID.Int64)
			parentID = &pid
		}
		items = append(items, v1.CategoryItem{
			ID:       cat.ID,
			Name:     cat.Name,
			ParentID: parentID,
			Icon:     cat.Icon,
		})
	}
	return &items, nil
}
