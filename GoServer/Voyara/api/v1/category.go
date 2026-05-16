package v1

import "github.com/gogf/gf/v2/frame/g"

type GetCategoriesReq struct {
	g.Meta `path:"/voyara/categories" method:"get" summary:"List categories"`
}

type CategoryItem struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ParentID *int   `json:"parentId"`
	Icon     string `json:"icon"`
}

type GetCategoriesRes []CategoryItem
