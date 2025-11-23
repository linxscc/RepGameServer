// 放入项目根目录的 main.go
package main

import (
	"GoServer/internal/controller"
	tcpserver "GoServer/tcpgameserver"

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
			"http://zspersonaldomain.it.com",
			"https://zspersonaldomain.it.com",
			"http://www.zspersonaldomain.it.com",  // 带 www
			"https://www.zspersonaldomain.it.com", // 带 www (HTTPS)
			"http://www.zspersonaldomain.it.com",  // 带 www.it
			"https://www.zspersonaldomain.it.com", // 带 www.it (HTTPS)
			"http://13.237.148.137",               // AWS EC2 公网IP
			"http://localhost:3000",               // 本地开发
			"http://localhost:5173",               // Vite 开发服务器
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
		group.Bind(
			new(controller.Normal),
			new(controller.ProductDocs),
		)
	})

	s.Run()
}
