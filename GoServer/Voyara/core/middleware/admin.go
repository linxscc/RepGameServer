package middleware

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// AdminOnly rejects non-admin users. Must be placed after Auth middleware
// which sets "userRole" in the request context.
func AdminOnly(r *ghttp.Request) {
	role, _ := r.Context().Value("userRole").(string)
	if role != "admin" {
		r.Response.WriteStatus(403)
		r.Response.WriteJson(g.Map{"code": 403, "message": "Forbidden: Admin access required"})
		return
	}
	r.Middleware.Next()
}
