package middleware

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// CSRF validates that mutating requests include a non-empty X-CSRF-Token header.
// The custom header itself provides CSRF protection for this Bearer-token SPA,
// since browsers cannot set custom headers cross-origin without CORS preflight,
// and our CORS policy restricts allowed origins.
func CSRF(exemptPaths ...string) ghttp.HandlerFunc {
	exempt := make(map[string]bool)
	for _, p := range exemptPaths {
		exempt[p] = true
	}

	return func(r *ghttp.Request) {
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			r.Middleware.Next()
			return
		}
		if exempt[r.URL.Path] {
			r.Middleware.Next()
			return
		}
		token := r.Header.Get("X-CSRF-Token")
		if token == "" {
			r.Response.WriteStatus(403)
			r.Response.WriteJson(g.Map{"code": 403, "message": "CSRF token required"})
			return
		}
		r.Middleware.Next()
	}
}
