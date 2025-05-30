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
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Bind(
			new(controller.Normal),
		)
	})
	s.Use(func(r *ghttp.Request) {
		r.Response.CORSDefault()
		r.Middleware.Next()
	})
	s.Run()
}
