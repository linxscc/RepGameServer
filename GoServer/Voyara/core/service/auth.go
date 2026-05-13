package service

import (
	"GoServer/Voyara/core/model"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret []byte

type JWTClaims struct {
	UserID int    `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func InitJWT(secret []byte) {
	jwtSecret = secret
}

func MakeAccessToken(userID int, email, role string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "voyara",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func MakeRefreshToken(userID int) (string, string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "voyara-refresh",
		Subject:   fmt.Sprintf("%d", userID),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}
	// Return both the token and its SHA-256 hash for storage
	hash := sha256.Sum256([]byte(tokenStr))
	return tokenStr, hex.EncodeToString(hash[:]), nil
}

func ParseAccessToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// IsLegacySHA256Hash checks if the stored hash is the old SHA-256 format (64 hex chars)
func IsLegacySHA256Hash(hash string) bool {
	if len(hash) != 64 {
		return false
	}
	_, err := hex.DecodeString(hash)
	return err == nil
}

// UpgradeLegacyPassword hashes a legacy SHA-256 password with bcrypt
func UpgradeLegacyPassword(password, oldHash string) (string, error) {
	legacySalt := "voyara-salt"
	legacyHash := sha256.Sum256([]byte(password + legacySalt))
	if hex.EncodeToString(legacyHash[:]) != oldHash {
		return "", errors.New("password doesn't match legacy hash")
	}
	return HashPassword(password)
}

func GetUserByEmail(email string) (*model.User, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var u model.User
	var emailVerifiedAt, lockedUntil sql.NullTime
	err = db.QueryRow(`
		SELECT id, email, password_hash, password_hash_method, name,
		       COALESCE(phone,''), COALESCE(country,''), COALESCE(preferred_lang,'en'),
		       email_verified_at, role, login_attempts, locked_until,
		       COALESCE(created_at,'')
		FROM voyara_users WHERE email = ?`, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.PasswordHashMethod, &u.Name,
			&u.Phone, &u.Country, &u.PreferredLang,
			&emailVerifiedAt, &u.Role, &u.LoginAttempts, &lockedUntil,
			&u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %v", err)
	}
	if emailVerifiedAt.Valid {
		u.EmailVerifiedAt = emailVerifiedAt
	}
	if lockedUntil.Valid {
		u.LockedUntil = lockedUntil
	}
	return &u, nil
}

func CreateUser(email, passwordHash, name string) (*model.User, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	res, err := db.Exec(`
		INSERT INTO voyara_users (email, password_hash, password_hash_method, name, role)
		VALUES (?, ?, 'bcrypt', ?, 'user')`, email, passwordHash, name)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") || strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, errors.New("email already registered")
		}
		return nil, fmt.Errorf("insert user: %v", err)
	}
	id, _ := res.LastInsertId()
	return &model.User{
		ID:    int(id),
		Email: email,
		Name:  name,
		Role:  "user",
	}, nil
}

func RecordLoginAttempt(email string, success bool, ip string) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	if success {
		_, err = db.Exec(`
			UPDATE voyara_users SET
			  login_attempts = 0,
			  locked_until = NULL,
			  last_login_at = NOW(),
			  last_login_ip = ?
			WHERE email = ?`, ip, email)
	} else {
		_, err = db.Exec(`
			UPDATE voyara_users SET
			  login_attempts = login_attempts + 1,
			  locked_until = IF(login_attempts + 1 >= 5, DATE_ADD(NOW(), INTERVAL 30 MINUTE), NULL)
			WHERE email = ?`, email)
	}
	return err
}

func StoreRefreshToken(userID int, tokenHash string, expiresAt time.Time) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`
		INSERT INTO voyara_refresh_tokens (user_id, token_hash, expires_at)
		VALUES (?, ?, ?)`, userID, tokenHash, expiresAt)
	return err
}

func ValidateRefreshToken(tokenHash string) (int, error) {
	db, err := GetDB()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var userID int
	var expiresAt time.Time
	err = db.QueryRow(`
		SELECT user_id, expires_at FROM voyara_refresh_tokens
		WHERE token_hash = ? AND revoked = 0`, tokenHash).
		Scan(&userID, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, errors.New("invalid refresh token")
	}
	if err != nil {
		return 0, fmt.Errorf("query refresh token: %v", err)
	}
	if time.Now().After(expiresAt) {
		return 0, errors.New("refresh token expired")
	}
	return userID, nil
}
