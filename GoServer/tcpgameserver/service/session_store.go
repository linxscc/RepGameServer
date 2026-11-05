package service

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

// SessionTTL 会话令牌有效期。令牌只用于断线重连时恢复身份。
const SessionTTL = 24 * time.Hour

type session struct {
	username  string
	expiresAt time.Time
}

// SessionStore 内存中的会话令牌存储（服务器重启后需重新登录）。
type SessionStore struct {
	mu     sync.Mutex
	byTok  map[string]session
	byUser map[string]string
	now    func() time.Time
}

var (
	sessionStoreOnce sync.Once
	sessionStore     *SessionStore
)

// GetSessionStore 返回全局会话存储。
func GetSessionStore() *SessionStore {
	sessionStoreOnce.Do(func() { sessionStore = NewSessionStore(time.Now) })
	return sessionStore
}

// NewSessionStore 创建会话存储；now 便于测试注入时间。
func NewSessionStore(now func() time.Time) *SessionStore {
	return &SessionStore{
		byTok:  make(map[string]session),
		byUser: make(map[string]string),
		now:    now,
	}
}

// Issue 为用户签发新令牌，同时作废该用户之前的令牌（单设备在线）。
func (s *SessionStore) Issue(username string) (token string, expiresAt time.Time, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", time.Time{}, err
	}
	token = base64.RawURLEncoding.EncodeToString(buf)
	expiresAt = s.now().Add(SessionTTL)

	s.mu.Lock()
	defer s.mu.Unlock()
	if old, ok := s.byUser[username]; ok {
		delete(s.byTok, old)
	}
	s.byTok[token] = session{username: username, expiresAt: expiresAt}
	s.byUser[username] = token
	return token, expiresAt, nil
}

// Resolve 返回令牌对应的用户名；令牌不存在或过期时 ok=false。
func (s *SessionStore) Resolve(token string) (username string, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, exists := s.byTok[token]
	if !exists {
		return "", false
	}
	if s.now().After(sess.expiresAt) {
		delete(s.byTok, token)
		if s.byUser[sess.username] == token {
			delete(s.byUser, sess.username)
		}
		return "", false
	}
	return sess.username, true
}
