package controller

import (
	v1 "GoServer/Voyara/api/v1"
	"GoServer/Voyara/core/service"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type Auth struct{}

func (c *Auth) SendVerificationCode(ctx context.Context, req *v1.SendVerificationCodeReq) (res *v1.MessageRes, err error) {
	if req.Purpose == "register" {
		existing, err := service.GetUserByEmail(req.Email)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, fmt.Errorf("email already registered")
		}
	}

	code, err := service.SaveVerificationCode(req.Email, req.Purpose)
	if err != nil {
		g.Log().Errorf(ctx, "SaveVerificationCode error: %v", err)
		return nil, err
	}

	if err := service.SendVerificationCode(req.Email, code, req.Purpose); err != nil {
		g.Log().Errorf(ctx, "SendVerificationCode error: %v", err)
		return nil, err
	}

	return &v1.MessageRes{Message: "Verification code sent"}, nil
}

func (c *Auth) Register(ctx context.Context, req *v1.RegisterReq) (res *v1.AuthRes, err error) {
	valid, err := service.VerifyCode(req.Email, req.Code, "register")
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("invalid or expired verification code")
	}

	existing, err := service.GetUserByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("email already registered")
	}

	hash, err := service.HashPassword(req.Password)
	if err != nil {
		g.Log().Errorf(ctx, "HashPassword error: %v", err)
		return nil, err
	}

	user, err := service.CreateUser(req.Email, hash, req.Name)
	if err != nil {
		return nil, err
	}

	_ = service.MarkEmailVerified(req.Email)

	token, err := service.MakeAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshHash, err := service.MakeRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}
	_ = service.StoreRefreshToken(user.ID, refreshHash, time.Now().Add(7*24*time.Hour))

	return &v1.AuthRes{
		Token:        token,
		RefreshToken: refreshToken,
		User: v1.UserInfo{
			ID:            user.ID,
			Email:         user.Email,
			Name:          user.Name,
			Role:          user.Role,
			EmailVerified: true,
		},
	}, nil
}

func (c *Auth) Login(ctx context.Context, req *v1.LoginReq) (res *v1.AuthRes, err error) {
	r := g.RequestFromCtx(ctx)
	ip := ""
	if r != nil {
		ip = r.GetClientIp()
	}

	user, err := service.GetUserByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if user.LockedUntil.Valid && user.LockedUntil.Time.After(time.Now()) {
		return nil, fmt.Errorf("account locked due to too many login attempts, try again later")
	}

	passwordValid := false
	if user.PasswordHashMethod == "bcrypt" {
		passwordValid = service.CheckPassword(req.Password, user.PasswordHash)
	} else {
		// Legacy SHA-256 password upgrade
		legacySalt := "voyara-salt"
		legacyHash := sha256.Sum256([]byte(req.Password + legacySalt))
		if hex.EncodeToString(legacyHash[:]) == user.PasswordHash {
			newHash, upgradeErr := service.HashPassword(req.Password)
			if upgradeErr == nil {
				upgradeUserPassword(user.ID, newHash)
				user.PasswordHashMethod = "bcrypt"
			}
			passwordValid = true
		}
	}

	if !passwordValid {
		_ = service.RecordLoginAttempt(req.Email, false, ip)
		return nil, fmt.Errorf("invalid email or password")
	}

	_ = service.RecordLoginAttempt(req.Email, true, ip)

	token, err := service.MakeAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}
	refreshToken, refreshHash, err := service.MakeRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}
	_ = service.StoreRefreshToken(user.ID, refreshHash, time.Now().Add(7*24*time.Hour))

	return &v1.AuthRes{
		Token:        token,
		RefreshToken: refreshToken,
		User: v1.UserInfo{
			ID:            user.ID,
			Email:         user.Email,
			Name:          user.Name,
			Role:          user.Role,
			EmailVerified: user.EmailVerifiedAt.Valid,
		},
	}, nil
}

func (c *Auth) RefreshToken(ctx context.Context, req *v1.RefreshTokenReq) (res *v1.RefreshTokenRes, err error) {
	// Hash the incoming token to look up in DB
	h := sha256.Sum256([]byte(req.RefreshToken))
	tokenHash := hex.EncodeToString(h[:])

	userID, err := service.ValidateRefreshToken(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	db, dbErr := service.GetDB()
	if dbErr != nil {
		return nil, dbErr
	}
	defer db.Close()

	var email, role string
	err = db.QueryRow(`SELECT email, role FROM voyara_users WHERE id = ?`, userID).Scan(&email, &role)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	newToken, err := service.MakeAccessToken(userID, email, role)
	if err != nil {
		return nil, err
	}
	newRefreshToken, newRefreshHash, err := service.MakeRefreshToken(userID)
	if err != nil {
		return nil, err
	}
	_ = service.StoreRefreshToken(userID, newRefreshHash, time.Now().Add(7*24*time.Hour))

	return &v1.RefreshTokenRes{
		Token:        newToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (c *Auth) ForgotPassword(ctx context.Context, req *v1.ForgotPasswordReq) (res *v1.MessageRes, err error) {
	user, err := service.GetUserByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &v1.MessageRes{Message: "If the email is registered, a reset code has been sent"}, nil
	}

	code, err := service.SaveVerificationCode(req.Email, "reset_password")
	if err != nil {
		return nil, err
	}

	if err := service.SendPasswordResetEmail(req.Email, code); err != nil {
		return nil, err
	}

	return &v1.MessageRes{Message: "If the email is registered, a reset code has been sent"}, nil
}

func (c *Auth) ResetPassword(ctx context.Context, req *v1.ResetPasswordReq) (res *v1.MessageRes, err error) {
	valid, err := service.VerifyCode(req.Email, req.Code, "reset_password")
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("invalid or expired verification code")
	}

	hash, err := service.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	db, dbErr := service.GetDB()
	if dbErr != nil {
		return nil, dbErr
	}
	defer db.Close()

	_, err = db.Exec(`UPDATE voyara_users SET password_hash = ?, password_hash_method = 'bcrypt' WHERE email = ?`, hash, req.Email)
	if err != nil {
		return nil, fmt.Errorf("reset password: %v", err)
	}

	return &v1.MessageRes{Message: "Password reset successfully"}, nil
}

func (c *Auth) ChangePassword(ctx context.Context, req *v1.ChangePasswordReq) (res *v1.MessageRes, err error) {
	userID := ctx.Value("userID").(int)

	db, dbErr := service.GetDB()
	if dbErr != nil {
		return nil, dbErr
	}
	defer db.Close()

	var currentHash, hashMethod string
	err = db.QueryRow(`SELECT password_hash, password_hash_method FROM voyara_users WHERE id = ?`, userID).
		Scan(&currentHash, &hashMethod)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if hashMethod == "bcrypt" {
		if !service.CheckPassword(req.OldPassword, currentHash) {
			return nil, fmt.Errorf("current password is incorrect")
		}
	} else {
		legacyHash := sha256.Sum256([]byte(req.OldPassword + "voyara-salt"))
		if hex.EncodeToString(legacyHash[:]) != currentHash {
			return nil, fmt.Errorf("current password is incorrect")
		}
	}

	newHash, err := service.HashPassword(req.NewPassword)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`UPDATE voyara_users SET password_hash = ?, password_hash_method = 'bcrypt' WHERE id = ?`, newHash, userID)
	if err != nil {
		return nil, fmt.Errorf("change password: %v", err)
	}

	return &v1.MessageRes{Message: "Password changed successfully"}, nil
}

// ── helpers ──

func upgradeUserPassword(userID int, hash string) {
	db, err := service.GetDB()
	if err != nil {
		return
	}
	defer db.Close()
	_, _ = db.Exec(`UPDATE voyara_users SET password_hash = ?, password_hash_method = 'bcrypt' WHERE id = ?`, hash, userID)
}

func init() {
	// Generate a random JWT secret at startup
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		g.Log().Fatalf(context.Background(), "Failed to generate JWT secret: %v", err)
	}
	service.InitJWT(b)
}
