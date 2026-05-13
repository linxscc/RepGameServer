package v1

import "github.com/gogf/gf/v2/frame/g"

type GetCartReq struct {
	g.Meta `path:"/voyara/cart" method:"get" summary:"Cart items" middleware:"auth"`
}

type AddCartReq struct {
	g.Meta    `path:"/voyara/cart" method:"post" summary:"Add to cart" middleware:"auth"`
	ProductID int `json:"productId" v:"required"`
	Quantity  int `json:"quantity" v:"required|min:1"`
}

type UpdateCartReq struct {
	g.Meta   `path:"/voyara/cart/:id" method:"put" summary:"Update cart item" middleware:"auth"`
	ID       int `json:"id" in:"path"`
	Quantity int `json:"quantity" v:"required|min:1"`
}

type ToggleCartSelectReq struct {
	g.Meta   `path:"/voyara/cart/:id/select" method:"put" summary:"Toggle item selected" middleware:"auth"`
	ID       int  `json:"id" in:"path"`
	Selected bool `json:"selected"`
}

type SelectAllCartReq struct {
	g.Meta   `path:"/voyara/cart/select-all" method:"put" summary:"Select/deselect all" middleware:"auth"`
	Selected bool `json:"selected"`
}

type DeleteCartReq struct {
	g.Meta `path:"/voyara/cart/:id" method:"delete" summary:"Remove cart item" middleware:"auth"`
	ID     int `json:"id" in:"path"`
}

type ClearCartReq struct {
	g.Meta `path:"/voyara/cart" method:"delete" summary:"Clear cart" middleware:"auth"`
}

type CartItemRes struct {
	ID             int     `json:"id"`
	ProductID      int     `json:"productId"`
	ProductTitle   string  `json:"productTitle"`
	ProductPrice   float64 `json:"productPrice"`
	ProductImage   string  `json:"productImage"`
	Quantity       int     `json:"quantity"`
	Selected       bool    `json:"selected"`
	SellerID       int     `json:"sellerId"`
	SellerShopName string  `json:"sellerShopName"`
	Available      bool    `json:"available"`
}

type GetCartRes struct {
	Items []CartItemRes `json:"items"`
	Count int           `json:"count"`
}

type AddCartRes struct {
	Message string `json:"message"`
}

type UpdateCartRes = CartItemRes

type CartCountRes struct {
	Count int `json:"count"`
}
