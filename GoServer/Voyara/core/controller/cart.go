package controller

import (
	v1 "GoServer/Voyara/api/v1"
	"GoServer/Voyara/core/service"
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

type Cart struct{}

func (c *Cart) GetCart(ctx context.Context, req *v1.GetCartReq) (res *v1.GetCartRes, err error) {
	userID := ctx.Value("userID").(int)
	items, err := service.GetCart(userID)
	if err != nil {
		g.Log().Errorf(ctx, "GetCart error: %v", err)
		return nil, err
	}
	itemRes := make([]v1.CartItemRes, 0, len(items))
	for _, item := range items {
		itemRes = append(itemRes, v1.CartItemRes{
			ID:             item.ID,
			ProductID:      item.ProductID,
			ProductTitle:   item.ProductTitle,
			ProductPrice:   item.ProductPrice,
			ProductImage:   item.ProductImage,
			Quantity:       item.Quantity,
			Selected:       item.Selected,
			SellerID:       item.SellerID,
			SellerShopName: item.SellerShopName,
			Available:      item.ProductStatus == "active",
		})
	}
	return &v1.GetCartRes{Items: itemRes, Count: len(itemRes)}, nil
}

func (c *Cart) AddCart(ctx context.Context, req *v1.AddCartReq) (res *v1.AddCartRes, err error) {
	userID := ctx.Value("userID").(int)
	if err := service.AddToCart(userID, req.ProductID, req.Quantity); err != nil {
		g.Log().Errorf(ctx, "AddCart error: %v", err)
		return nil, err
	}
	return &v1.AddCartRes{Message: "Added to cart"}, nil
}

func (c *Cart) UpdateCart(ctx context.Context, req *v1.UpdateCartReq) (res *v1.CartItemRes, err error) {
	userID := ctx.Value("userID").(int)
	if err := service.UpdateCartItemQuantity(userID, req.ID, req.Quantity); err != nil {
		g.Log().Errorf(ctx, "UpdateCart error: %v", err)
		return nil, err
	}
	return &v1.CartItemRes{ID: req.ID, Quantity: req.Quantity}, nil
}

func (c *Cart) ToggleCartSelect(ctx context.Context, req *v1.ToggleCartSelectReq) (res *v1.MessageRes, err error) {
	userID := ctx.Value("userID").(int)
	if err := service.ToggleCartItemSelected(userID, req.ID, req.Selected); err != nil {
		g.Log().Errorf(ctx, "ToggleCartSelect error: %v", err)
		return nil, err
	}
	return &v1.MessageRes{Message: "Updated"}, nil
}

func (c *Cart) SelectAll(ctx context.Context, req *v1.SelectAllCartReq) (res *v1.MessageRes, err error) {
	userID := ctx.Value("userID").(int)
	if err := service.SelectAllCartItems(userID, req.Selected); err != nil {
		g.Log().Errorf(ctx, "SelectAll error: %v", err)
		return nil, err
	}
	return &v1.MessageRes{Message: "Updated"}, nil
}

func (c *Cart) DeleteCart(ctx context.Context, req *v1.DeleteCartReq) (res *v1.MessageRes, err error) {
	userID := ctx.Value("userID").(int)
	if err := service.RemoveCartItem(userID, req.ID); err != nil {
		g.Log().Errorf(ctx, "DeleteCart error: %v", err)
		return nil, err
	}
	return &v1.MessageRes{Message: "Removed"}, nil
}

func (c *Cart) ClearCart(ctx context.Context, req *v1.ClearCartReq) (res *v1.MessageRes, err error) {
	userID := ctx.Value("userID").(int)
	if err := service.ClearCart(userID); err != nil {
		g.Log().Errorf(ctx, "ClearCart error: %v", err)
		return nil, err
	}
	return &v1.MessageRes{Message: "Cart cleared"}, nil
}
