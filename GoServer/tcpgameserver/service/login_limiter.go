package service

import (
	"sync"
	"time"

	"GoServer/tcpgameserver/config"
)

// 登录/注册限流：按固定 1 分钟窗口计数，超过上限后在窗口剩余时间内直接拒绝。
// 每次尝试都要做一次 bcrypt（数百毫秒 CPU），不限流的话少量并发就能占满 CPU，也给撞库留了口子。

// RateLimiter 按 key 计数的固定窗口限流器
type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	now     func() time.Time
	buckets map[string]*rateBucket
	calls   int
}

type rateBucket struct {
	start time.Time
	count int
}

// NewRateLimiter 每个 key 在 window 内最多允许 limit 次；now 便于测试注入时间
func NewRateLimiter(limit int, window time.Duration, now func() time.Time) *RateLimiter {
	return &RateLimiter{limit: limit, window: window, now: now, buckets: make(map[string]*rateBucket)}
}

// Allow 记录一次尝试，返回是否放行
func (l *RateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	l.calls++
	if l.calls%1024 == 0 {
		l.sweep(now)
	}

	b, ok := l.buckets[key]
	if !ok || now.Sub(b.start) >= l.window {
		l.buckets[key] = &rateBucket{start: now, count: 1}
		return true
	}
	if b.count >= l.limit {
		return false
	}
	b.count++
	return true
}

// sweep 清理已过期的窗口，避免大量不同 IP/用户名让内存持续增长
func (l *RateLimiter) sweep(now time.Time) {
	for k, b := range l.buckets {
		if now.Sub(b.start) >= l.window {
			delete(l.buckets, k)
		}
	}
}

var (
	limiterOnce sync.Once
	ipLimiter   *RateLimiter
	userLimiter *RateLimiter
)

// AllowAuthAttempt 一次登录/注册尝试是否放行（network.yaml 的 auth.login_attempts_*）。
// username 为空时只按 IP 计数（例如注册）。
func AllowAuthAttempt(ip, username string) bool {
	limiterOnce.Do(func() {
		auth := config.Game().Network.Auth
		ipLimiter = NewRateLimiter(auth.LoginAttemptsPerIPPerMinute, time.Minute, time.Now)
		userLimiter = NewRateLimiter(auth.LoginAttemptsPerUserPerMinute, time.Minute, time.Now)
	})
	if !ipLimiter.Allow(ip) {
		return false
	}
	return username == "" || userLimiter.Allow(username)
}
