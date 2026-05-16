package controller

import (
	v1 "GoServer/Voyara/api/v1"
	"GoServer/Voyara/core/model"
	"GoServer/Voyara/core/service"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

type Product struct{}

func (c *Product) GetProducts(ctx context.Context, req *v1.GetProductsReq) (res *v1.GetProductsRes, err error) {
	result, err := service.GetProducts(service.ProductFilter{
		Category:  req.Category,
		Condition: req.Condition,
		MinPrice:  req.MinPrice,
		MaxPrice:  req.MaxPrice,
		Search:    req.Search,
		Page:      req.Page,
		PageSize:  req.PageSize,
	})
	if err != nil {
		g.Log().Errorf(ctx, "GetProducts error: %v", err)
		return nil, err
	}

	items := make([]v1.ProductItem, 0, len(result.Items))
	for _, p := range result.Items {
		item := toProductItem(p)
		items = append(items, item)
	}
	return &v1.GetProductsRes{
		Items:    items,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}

func (c *Product) GetProduct(ctx context.Context, req *v1.GetProductReq) (res *v1.GetProductRes, err error) {
	p, err := service.GetProductByID(req.ID)
	if err != nil {
		g.Log().Errorf(ctx, "GetProduct error: %v", err)
		return nil, err
	}
	item := toProductItem(*p)
	return &item, nil
}

// resolveUserID extracts the authenticated user ID from context, or falls back
// to parsing the JWT from Authorization header / Cookie.
func resolveUserID(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value("userID").(int)
	if ok {
		return userID, true
	}
	r := g.RequestFromCtx(ctx)
	tokenStr := ""
	if cookie, err := r.Request.Cookie("voyara_token"); err == nil && cookie != nil {
		tokenStr = cookie.Value
	}
	if tokenStr == "" {
		header := r.Header.Get("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			tokenStr = strings.TrimPrefix(header, "Bearer ")
		}
	}
	if tokenStr == "" {
		return 0, false
	}
	claims, err := service.ParseAccessToken(tokenStr)
	if err != nil {
		return 0, false
	}
	return claims.UserID, true
}

func (c *Product) CreateProduct(ctx context.Context, req *v1.CreateProductReq) (res *v1.CreateProductRes, err error) {
	userID, ok := resolveUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("unauthorized")
	}
	sellerID, err := service.EnsureSeller(userID, "My Shop", "")
	if err != nil {
		return nil, err
	}

	// Upload images from multipart form data
	var imageURLs []string
	r := g.RequestFromCtx(ctx)
	if files := r.GetMultipartFiles("images"); len(files) > 0 {
		if len(files) > 5 {
			return nil, fmt.Errorf("too many images (max 5)")
		}
		for _, f := range files {
			file, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("open uploaded file: %v", err)
			}
			data, err := io.ReadAll(file)
			file.Close()
			if err != nil {
				return nil, fmt.Errorf("read uploaded file: %v", err)
			}
			if err := service.ValidateImageFile(f.Filename, data); err != nil {
				return nil, err
			}
			url, err := service.UploadToS3(data, f.Filename)
			if err != nil {
				g.Log().Errorf(ctx, "S3 upload failed: %v", err)
				return nil, fmt.Errorf("image upload failed")
			}
			imageURLs = append(imageURLs, url)
		}
	}

	p, err := service.CreateProduct(sellerID, req.Title, req.Description, service.DollarsToCents(req.Price), req.Category, req.Condition, imageURLs)
	if err != nil {
		g.Log().Errorf(ctx, "CreateProduct error: %v", err)
		return nil, err
	}
	item := toProductItem(*p)
	return &item, nil
}

func (c *Product) UpdateProduct(ctx context.Context, req *v1.UpdateProductReq) (res *v1.CreateProductRes, err error) {
	userID, ok := resolveUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("unauthorized")
	}
	sellerID, err := service.GetSellerIDByUserID(userID)
	if err != nil || sellerID == 0 {
		return nil, fmt.Errorf("seller not found")
	}

	var priceCents *int64
	if req.Price != nil {
		c := service.DollarsToCents(*req.Price)
		priceCents = &c
	}
	p, err := service.UpdateProduct(req.ID, sellerID, req.Title, req.Description, priceCents, req.Category, req.Condition)
	if err != nil {
		g.Log().Errorf(ctx, "UpdateProduct error: %v", err)
		return nil, err
	}
	item := toProductItem(*p)
	return &item, nil
}

func toProductItem(p model.Product) v1.ProductItem {
	desc := ""
	if p.Description.Valid {
		desc = p.Description.String
	}
	var images []string
	if p.Images.Valid {
		json.Unmarshal([]byte(p.Images.String), &images)
	}
	if images == nil {
		images = []string{}
	}
	return v1.ProductItem{
		ID:          p.ID,
		SellerID:    p.SellerID,
		SellerName:  p.SellerName,
		ShopName:    p.ShopName,
		Title:       p.Title,
		Description: desc,
		Price:       service.CentsToDollars(p.Price),
		Currency:    p.Currency,
		Category:    p.Category,
		Condition:   p.Condition,
		Images:      images,
		Status:      p.Status,
		CreatedAt:   p.CreatedAt,
	}
}

func (c *Product) GetSellerProducts(ctx context.Context, req *v1.GetSellerProductsReq) (res *v1.GetSellerProductsRes, err error) {
	userID, ok := resolveUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("unauthorized")
	}
	sellerID, err := service.GetSellerIDByUserID(userID)
	if err != nil || sellerID == 0 {
		return nil, fmt.Errorf("seller not found")
	}
	products, err := service.GetProductsBySeller(sellerID)
	if err != nil {
		g.Log().Errorf(ctx, "GetSellerProducts error: %v", err)
		return nil, err
	}
	items := make([]v1.ProductItem, 0, len(products))
	for _, p := range products {
		items = append(items, toProductItem(p))
	}
	return &v1.GetSellerProductsRes{Items: items}, nil
}
