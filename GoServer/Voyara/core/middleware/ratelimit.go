package middleware

import (
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
}

type visitor struct {
	count  int
	expiry time.Time
}

var rl = &rateLimiter{
	visitors: make(map[string]*visitor),
}

func init() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.mu.Lock()
			now := time.Now()
			for key, v := range rl.visitors {
				if now.After(v.expiry) {
					delete(rl.visitors, key)
				}
			}
			rl.mu.Unlock()
		}
	}()
}

func (rl *rateLimiter) allow(key string, limit int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[key]
	now := time.Now()

	if !exists || now.After(v.expiry) {
		rl.visitors[key] = &visitor{count: 1, expiry: now.Add(window)}
		return true
	}

	if v.count >= limit {
		return false
	}

	v.count++
	return true
}

func RateLimit(keyFn func(*ghttp.Request) string, limit int, window time.Duration) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		key := keyFn(r)
		if !rl.allow(key, limit, window) {
			r.Response.WriteStatus(429)
			r.Response.WriteJson(g.Map{
				"code":    429,
				"message": "Too many requests, please try again later",
			})
			return
		}
		r.Middleware.Next()
	}
}

func StrictRateLimit() ghttp.HandlerFunc {
	return RateLimit(func(r *ghttp.Request) string {
		return "strict:" + r.GetClientIp()
	}, 5, 15*time.Minute)
}

func DefaultRateLimit() ghttp.HandlerFunc {
	return RateLimit(func(r *ghttp.Request) string {
		return "default:" + r.GetClientIp()
	}, 100, 1*time.Minute)
}
