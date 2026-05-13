package controller

import (
	v1 "GoServer/Voyara/api/v1"
	"GoServer/Voyara/core/model"
	"GoServer/Voyara/core/service"
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
)

type Order struct{}

func (c *Order) Checkout(ctx context.Context, req *v1.CheckoutReq) (res *v1.CheckoutRes, err error) {
	userID := ctx.Value("userID").(int)

	order, err := service.Checkout(userID, req.ProductIDs, service.ShippingAddr{
		Name:    req.ShippingAddress.Name,
		Phone:   req.ShippingAddress.Phone,
		Country: req.ShippingAddress.Country,
		City:    req.ShippingAddress.City,
		Street:  req.ShippingAddress.Street,
		ZipCode: req.ShippingAddress.ZipCode,
	}, req.IdempotencyKey)
	if err != nil {
		g.Log().Errorf(ctx, "Checkout error: %v", err)
		return nil, err
	}
	item := toOrderItemRes(*order)
	return &item, nil
}

func (c *Order) GetOrders(ctx context.Context, req *v1.GetOrdersReq) (res *v1.GetOrdersRes, err error) {
	userID := ctx.Value("userID").(int)

	orders, err := service.GetOrdersByBuyer(userID)
	if err != nil {
		g.Log().Errorf(ctx, "GetOrders error: %v", err)
		return nil, err
	}

	items := make(v1.GetOrdersRes, 0, len(orders))
	for _, o := range orders {
		items = append(items, toOrderItemRes(o))
	}
	return &items, nil
}

func (c *Order) ShipOrder(ctx context.Context, req *v1.ShipOrderReq) (res *v1.OrderItemRes, err error) {
	userID := ctx.Value("userID").(int)
	sellerID, err := service.GetSellerIDByUserID(userID)
	if err != nil || sellerID == 0 {
		return nil, fmt.Errorf("seller not found")
	}

	if err := service.ShipOrder(req.ID, sellerID, req.TrackingNumber); err != nil {
		g.Log().Errorf(ctx, "ShipOrder error: %v", err)
		return nil, err
	}
	return &v1.OrderItemRes{ID: req.ID, TrackingNumber: req.TrackingNumber, ShippingStatus: "shipped"}, nil
}

func toOrderItemRes(o model.Order) v1.OrderItemRes {
	var addr v1.ShippingAddr
	if o.ShippingAddress.Valid {
		_ = json.Unmarshal([]byte(o.ShippingAddress.String), &addr)
	}
	items := make([]v1.OrderItem, 0, len(o.Items))
	for _, item := range o.Items {
		items = append(items, v1.OrderItem{
			ID:        item.ID,
			ProductID: item.ProductID,
			Title:     item.Title,
			Price:     item.Price,
			Quantity:  item.Quantity,
			Total:     item.Total,
			ImageURL:  item.ImageURL,
		})
	}
	return v1.OrderItemRes{
		ID:              o.ID,
		OrderNo:         o.OrderNo,
		BuyerID:         o.BuyerID,
		SellerID:        o.SellerID,
		ItemCount:       o.ItemCount,
		Amount:          o.Amount,
		Subtotal:        o.Subtotal,
		Currency:        o.Currency,
		PaymentStatus:   o.PaymentStatus,
		ShippingStatus:  o.ShippingStatus,
		TrackingNumber:  o.TrackingNumber,
		ShippingAddress: addr,
		CreatedAt:       o.CreatedAt,
		Items:           items,
	}
}
