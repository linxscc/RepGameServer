package service

import (
	"GoServer/tcpgameserver/config"
	"crypto/subtle"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidCredentials 用户不存在或密码错误（两者不区分，避免泄露用户是否存在）。
var ErrInvalidCredentials = errors.New("invalid username or password")

// 用户不存在时也做一次 bcrypt 比较，使响应耗时一致。
var (
	dummyHashOnce sync.Once
	dummyHash     []byte
)

func timingDummyHash() []byte {
	dummyHashOnce.Do(func() {
		dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcryptCost())
	})
	return dummyHash
}

// bcryptCost bcrypt 计算强度（network.yaml 的 auth.bcrypt_cost）
func bcryptCost() int {
	return config.Game().Network.Auth.BcryptCost
}

// 同时进行的 bcrypt 计算数量上限（network.yaml 的 auth.max_concurrent_hashes，0 = CPU 核数）。
// 超出的请求排队，保证登录高峰时游戏逻辑仍有 CPU 可用。
var (
	hashSlotsOnce sync.Once
	hashSlots     chan struct{}
)

func withHashSlot(fn func()) {
	hashSlotsOnce.Do(func() {
		n := config.Game().Network.Auth.MaxConcurrentHashes
		if n <= 0 {
			n = runtime.NumCPU()
		}
		hashSlots = make(chan struct{}, n)
	})
	hashSlots <- struct{}{}
	defer func() { <-hashSlots }()
	fn()
}

func compareHash(hash []byte, password string) (err error) {
	withHashSlot(func() { err = bcrypt.CompareHashAndPassword(hash, []byte(password)) })
	return err
}

// HashPassword 生成 bcrypt 哈希。
func HashPassword(password string) (string, error) {
	var hash []byte
	var err error
	withHashSlot(func() { hash, err = bcrypt.GenerateFromPassword([]byte(password), bcryptCost()) })
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// ValidateCredentialFormat 校验注册/登录时用户名和密码的基本格式（规则见 network.yaml 的 auth）。
// 返回的错误文本来自 messages.yaml，可直接回复给客户端。
func ValidateCredentialFormat(username, password string) error {
	auth := config.Game().Network.Auth
	if n := len([]rune(username)); n < auth.UsernameMinLength || n > auth.UsernameMaxLength {
		return errors.New(config.Msg(config.MsgUsernameLength, auth.UsernameMinLength, auth.UsernameMaxLength))
	}
	if strings.TrimSpace(username) != username {
		return errors.New(config.Msg(config.MsgUsernameSpaces))
	}
	if n := len(password); n < auth.PasswordMinLength || n > auth.PasswordMaxLength {
		return errors.New(config.Msg(config.MsgPasswordLength, auth.PasswordMinLength, auth.PasswordMaxLength))
	}
	return nil
}

// IsReservedUsername 机器人名称前缀保留给系统使用，真人玩家不能注册
func IsReservedUsername(username string) bool {
	return strings.HasPrefix(username, config.Game().Bot.NamePrefix)
}

// AuthenticateUser 校验用户名和密码。
// 返回 ErrInvalidCredentials 表示账号或密码错误；其它错误为服务器内部错误。
//
// 兼容旧数据：库中仍是明文密码的账号，验证通过后会自动升级为 bcrypt 哈希。
func AuthenticateUser(username, password string) error {
	user, err := GetUserAccount(username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			_ = compareHash(timingDummyHash(), password)
			return ErrInvalidCredentials
		}
		return err
	}

	if isBcryptHash(user.Password) {
		if compareHash([]byte(user.Password), password) != nil {
			return ErrInvalidCredentials
		}
		return nil
	}

	// 旧的明文密码
	if subtle.ConstantTimeCompare([]byte(user.Password), []byte(password)) != 1 {
		_ = compareHash(timingDummyHash(), password)
		return ErrInvalidCredentials
	}
	if hash, err := HashPassword(password); err == nil {
		_ = updatePasswordHash(username, hash) // 升级失败不影响本次登录，下次登录会再试
	}
	return nil
}

func isBcryptHash(s string) bool {
	return strings.HasPrefix(s, "$2a$") || strings.HasPrefix(s, "$2b$") || strings.HasPrefix(s, "$2y$")
}
