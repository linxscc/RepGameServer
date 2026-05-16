// 放入项目根目录的 main.go
package main

import (
	"GoServer/internal/controller"
	tcpserver "GoServer/tcpgameserver"
	voyaraController "GoServer/Voyara/core/controller"
	voyaraMiddleware "GoServer/Voyara/core/middleware"
	voyaraService "GoServer/Voyara/core/service"
	"log"
	"os"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/joho/godotenv"
)

// init() 在 Go 中的执行顺序:
//   1. 所有 import 包的 init() 函数
//   2. 本包的 init() 函数
//   3. main()
// 之前 service/auth.go 的 init() 会在本 init() 之前执行，
// 导致读不到 .env 中的 VOYARA_JWT_SECRET。
// 现在已移除 auth.go 的 init()，统一由 main() 显式调用 InitJWT 完成初始化。

func init() {
	envPath := os.Getenv("VOYARA_ENV_PATH")
	if envPath == "" {
		envPath = ".env"
	}
	if err := godotenv.Load(envPath); err != nil {
		log.Println("[voyara] No .env file found, using system environment variables")
	}
}

func main() {
	go tcpserver.StartTCPServer()

	voyaraService.StartPaymentTimeoutScheduler()

	// 初始化 JWT：必须在 godotenv.Load() 之后调用，
	// VOYARA_JWT_SECRET 未设置或不足 32 字符时服务启动失败。
	if err := voyaraService.InitJWT(os.Getenv("VOYARA_JWT_SECRET")); err != nil {
		log.Fatalf("[voyara] Failed to initialize JWT: %v", err)
	}

	// 初始化 Voyara 数据库连接池
	if err := voyaraService.InitDB(); err != nil {
		log.Fatalf("[voyara] Failed to initialize database: %v", err)
	}
	defer voyaraService.CloseDB()

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
		r.Response.Header().Set("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,User-Agent,Cookie,Authorization,X-Auth-Token,X-Requested-With,X-CSRF-Token")
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

	// ── Voyara Marketplace Public Endpoints (no auth) ──
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Bind(
			new(voyaraController.Auth),
			new(voyaraController.Product),
			new(voyaraController.Category),
		)
	})

	// ── Voyara Marketplace Protected Endpoints (auth + CSRF required) ──
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Middleware(voyaraMiddleware.Auth)
		group.Middleware(voyaraMiddleware.CSRF())
		group.Bind(
			new(voyaraController.Order),
			new(voyaraController.Cart),
			new(voyaraController.Payment),
		)
	})

	// ── Voyara Marketplace Admin Endpoints (auth + admin + CSRF) ──
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Middleware(voyaraMiddleware.Auth)
		group.Middleware(voyaraMiddleware.AdminOnly)
		group.Middleware(voyaraMiddleware.CSRF())
		group.Bind(
			new(voyaraController.Admin),
		)
	})

	// ── Voyara Payment Webhooks (no auth, raw request body) ──
	{
		pay := &voyaraController.Payment{}
		s.BindHandler("POST:/voyara/payment/stripe-webhook", pay.StripeWebhook)
		s.BindHandler("POST:/voyara/payment/paypal-webhook", pay.PayPalWebhook)
	}

	s.Run()
}
