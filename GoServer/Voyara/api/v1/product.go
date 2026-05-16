package v1

import "github.com/gogf/gf/v2/frame/g"

type GetProductsReq struct {
	g.Meta    `path:"/voyara/products" method:"get" summary:"List products"`
	Category  string `json:"category" dc:"appliance|vehicle|electronics|other"`
	Condition string `json:"condition" dc:"new|like_new|used|refurbished"`
	MinPrice  string `json:"minPrice"`
	MaxPrice  string `json:"maxPrice"`
	Search    string `json:"search"`
	Page      int    `json:"page" dc:"Page number, default 1"`
	PageSize  int    `json:"pageSize" dc:"Items per page, default 20, max 100"`
}

type GetProductReq struct {
	g.Meta `path:"/voyara/products/:id" method:"get" summary:"Get product"`
	ID     int `json:"id" in:"path"`
}

type CreateProductReq struct {
	g.Meta      `path:"/voyara/products" method:"post" summary:"Create product" middleware:"auth"`
	Title       string  `json:"title" v:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" v:"required|min:0"`
	Category    string  `json:"category" v:"required"`
	Condition   string  `json:"condition" v:"required"`
}

type UpdateProductReq struct {
	g.Meta      `path:"/voyara/products/:id" method:"put" summary:"Update product" middleware:"auth"`
	ID          int      `json:"id" in:"path"`
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	Category    *string  `json:"category"`
	Condition   *string  `json:"condition"`
}

type GetProductsRes struct {
	Items    []ProductItem `json:"items"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}

type ProductItem struct {
	ID          int      `json:"id"`
	SellerID    int      `json:"sellerId"`
	SellerName  string   `json:"sellerName,omitempty"`
	ShopName    string   `json:"shopName,omitempty"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Currency    string   `json:"currency"`
	Category    string   `json:"category"`
	Condition   string   `json:"condition"`
	Images      []string `json:"images"`
	Status      string   `json:"status"`
	CreatedAt   string   `json:"createdAt"`
}

type GetProductRes = ProductItem

type CreateProductRes = ProductItem

type GetSellerProductsReq struct {
	g.Meta `path:"/voyara/seller/products" method:"get" summary:"Get my products" middleware:"auth"`
}

type GetSellerProductsRes struct {
	Items []ProductItem `json:"items"`
}
