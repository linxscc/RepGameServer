package service

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = 12

	// 用户名/密码的长度约束（bcrypt 只使用前 72 字节）
	MinUsernameLength = 3
	MaxUsernameLength = 32
	MinPasswordLength = 6
	MaxPasswordLength = 72
)

// ErrInvalidCredentials 用户不存在或密码错误（两者不区分，避免泄露用户是否存在）。
var ErrInvalidCredentials = errors.New("invalid username or password")

// 用户不存在时也做一次 bcrypt 比较，使响应耗时一致。
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcryptCost)

// HashPassword 生成 bcrypt 哈希。
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// ValidateCredentialFormat 校验注册/登录时用户名和密码的基本格式。
func ValidateCredentialFormat(username, password string) error {
	if n := len([]rune(username)); n < MinUsernameLength || n > MaxUsernameLength {
		return fmt.Errorf("username length must be %d-%d", MinUsernameLength, MaxUsernameLength)
	}
	if strings.TrimSpace(username) != username {
		return errors.New("username must not start or end with spaces")
	}
	if n := len(password); n < MinPasswordLength || n > MaxPasswordLength {
		return fmt.Errorf("password length must be %d-%d bytes", MinPasswordLength, MaxPasswordLength)
	}
	return nil
}

// AuthenticateUser 校验用户名和密码。
// 返回 ErrInvalidCredentials 表示账号或密码错误；其它错误为服务器内部错误。
//
// 兼容旧数据：库中仍是明文密码的账号，验证通过后会自动升级为 bcrypt 哈希。
func AuthenticateUser(username, password string) error {
	user, err := GetUserAccount(username)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
			return ErrInvalidCredentials
		}
		return err
	}

	if isBcryptHash(user.Password) {
		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
			return ErrInvalidCredentials
		}
		return nil
	}

	// 旧的明文密码
	if subtle.ConstantTimeCompare([]byte(user.Password), []byte(password)) != 1 {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
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

func updatePasswordHash(username, hash string) error {
	db, err := GetDBConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("UPDATE UserAccount SET password = ? WHERE username = ?", hash, username)
	return err
}
