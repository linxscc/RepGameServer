package v1

import "github.com/gogf/gf/v2/frame/g"

type AdminGetProductsReq struct {
	g.Meta `path:"/voyara/admin/products" method:"get" summary:"Admin list products"`
	Status string `json:"status" dc:"Filter by: active|inactive|in_review"`
}

type AdminGetProductsRes struct {
	Items []ProductItem `json:"items"`
}

type AdminUpdateProductStatusReq struct {
	g.Meta `path:"/voyara/admin/products/:id/status" method:"put" summary:"Admin update product status"`
	ID     int    `json:"id" in:"path"`
	Status string `json:"status" v:"required|in:active,inactive,in_review"`
}
