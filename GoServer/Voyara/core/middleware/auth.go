package middleware

import (
	"GoServer/Voyara/core/service"
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

var publicPaths = []string{
	"POST /voyara/auth/login",
	"POST /voyara/auth/register",
	"POST /voyara/auth/send-verification",
	"POST /voyara/auth/refresh",
	"POST /voyara/auth/forgot-password",
	"POST /voyara/auth/reset-password",
	"POST /voyara/auth/verify-email",
	"POST /voyara/payment/stripe-webhook",
	"POST /voyara/payment/paypal-webhook",
}

func isPublic(method, path string) bool {
	if method == "GET" && (strings.HasPrefix(path, "/voyara/products") ||
		path == "/voyara/categories" ||
		path == "/voyara/brands") {
		return true
	}
	sig := method + " " + path
	for _, p := range publicPaths {
		if sig == p {
			return true
		}
	}
	return false
}

func Auth(r *ghttp.Request) {
	if isPublic(r.Method, r.URL.Path) {
		r.Middleware.Next()
		return
	}

	tokenStr := ""
	if cookie, err := r.Request.Cookie("voyara_token"); err == nil && cookie != nil {
		tokenStr = cookie.Value
	}
	if tokenStr == "" {
		header := r.Header.Get("Authorization")
		if header != "" && strings.HasPrefix(header, "Bearer ") {
			tokenStr = strings.TrimPrefix(header, "Bearer ")
		}
	}
	if tokenStr == "" {
		r.Response.WriteStatus(401)
		r.Response.WriteJson(g.Map{"code": 401, "message": "Authentication required"})
		return
	}
	claims, err := service.ParseAccessToken(tokenStr)
	if err != nil {
		r.Response.WriteStatus(401)
		r.Response.WriteJson(g.Map{"code": 401, "message": "Invalid or expired token"})
		return
	}

	ctx := context.WithValue(r.Context(), "userID", claims.UserID)
	ctx = context.WithValue(ctx, "userRole", claims.Role)
	ctx = context.WithValue(ctx, "userEmail", claims.Email)
	r.SetCtx(ctx)
	r.Middleware.Next()
}
