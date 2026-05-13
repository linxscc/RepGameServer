// 放入项目根目录的 main.go
package main

import (
	"GoServer/internal/controller"
	tcpserver "GoServer/tcpgameserver"
	voyaraController "GoServer/Voyara/core/controller"
	voyaraMiddleware "GoServer/Voyara/core/middleware"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func main() {
	go tcpserver.StartTCPServer()

	s := g.Server()

	// 配置 CORS - 必须在路由之前设置
	// 只允许指定的域名访问
	s.Use(func(r *ghttp.Request) {
		origin := r.Header.Get("Origin")
		// 允许的域名列表
		allowedOrigins := []string{
			"http://zsdimain.site",
			"https://zsdimain.site",
			"http://www.zsdimain.site",
			"https://www.zsdimain.site",
			"http://13.237.148.137",
			"https://13.237.148.137",
			"http://localhost:3000",
			"http://localhost:5173",
		}

		// 检查请求来源是否在允许列表中
		isAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				r.Response.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		// 如果来源不在允许列表中，拒绝请求
		if !isAllowed && origin != "" {
			g.Log().Warningf(r.Context(), "CORS: 拒绝来自未授权域名的请求: %s", origin)
			r.Response.WriteStatus(403)
			r.Response.WriteJson(g.Map{
				"code":    403,
				"message": "Access denied: Origin not allowed",
			})
			return
		}

		r.Response.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,PATCH,HEAD,CONNECT,OPTIONS,TRACE")
		r.Response.Header().Set("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,User-Agent,Cookie,Authorization,X-Auth-Token,X-Requested-With")
		r.Response.Header().Set("Access-Control-Allow-Credentials", "true")
		r.Response.Header().Set("Access-Control-Max-Age", "3600")

		// 处理 OPTIONS 预检请求
		if r.Method == "OPTIONS" {
			r.Response.WriteStatus(200)
			return
		}

		r.Middleware.Next()
	})

	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Bind(
			new(controller.Normal),
			new(controller.ProductDocs),
			new(controller.Download),
			new(controller.Profile),
			new(controller.WorkExperience),
		)
	})

	// ── Voyara Marketplace Public Endpoints (no auth/CSRF for auth endpoints) ──
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Bind(
			new(voyaraController.Auth),
		)
	})

	// ── Voyara Marketplace Protected Endpoints ──
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Middleware(voyaraMiddleware.Auth)
		group.Middleware(voyaraMiddleware.CSRF())
		group.Bind(
			new(voyaraController.Product),
			new(voyaraController.Category),
			new(voyaraController.Order),
		)
	})

	s.Run()
}
