package controller

import (
	"context"
	"fmt"

	v1 "GoServer/Voyara/api/v1"
	"GoServer/Voyara/core/service"

	"github.com/gogf/gf/v2/frame/g"
)

type Admin struct{}

func (c *Admin) GetProducts(ctx context.Context, req *v1.AdminGetProductsReq) (res *v1.AdminGetProductsRes, err error) {
	products, err := service.GetAllProducts(req.Status)
	if err != nil {
		g.Log().Errorf(ctx, "AdminGetProducts error: %v", err)
		return nil, err
	}
	items := make([]v1.ProductItem, 0, len(products))
	for _, p := range products {
		items = append(items, toProductItem(p))
	}
	return &v1.AdminGetProductsRes{Items: items}, nil
}

func (c *Admin) UpdateProductStatus(ctx context.Context, req *v1.AdminUpdateProductStatusReq) (res *v1.MessageRes, err error) {
	if err := service.UpdateProductStatus(req.ID, req.Status); err != nil {
		g.Log().Errorf(ctx, "AdminUpdateProductStatus error: %v", err)
		return nil, err
	}
	return &v1.MessageRes{Message: fmt.Sprintf("Product status updated to %s", req.Status)}, nil
}
