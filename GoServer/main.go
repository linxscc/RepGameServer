// 放入项目根目录的 main.go
package main

import (
	"hello/internal/controller"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func main() {
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
