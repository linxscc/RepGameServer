package v1

import "github.com/gogf/gf/v2/frame/g"

type RegisterReq struct {
	g.Meta    `path:"/voyara/auth/register" method:"post" summary:"Register"`
	Email     string `json:"email" v:"required|email|length:5,255"`
	Password  string `json:"password" v:"required|length:8,64"`
	Name      string `json:"name" v:"required|length:1,100"`
	Code      string `json:"code" v:"required|length:6,6"`
}

type LoginReq struct {
	g.Meta   `path:"/voyara/auth/login" method:"post" summary:"Login"`
	Email    string `json:"email" v:"required|email"`
	Password string `json:"password" v:"required"`
}

type SendVerificationCodeReq struct {
	g.Meta  `path:"/voyara/auth/send-verification" method:"post" summary:"Send verification code"`
	Email   string `json:"email" v:"required|email"`
	Purpose string `json:"purpose" v:"required|in:register,reset_password"`
}

type VerifyEmailReq struct {
	g.Meta  `path:"/voyara/auth/verify-email" method:"post" summary:"Verify email with code"`
	Email   string `json:"email" v:"required|email"`
	Code    string `json:"code" v:"required|length:6,6"`
}

type RefreshTokenReq struct {
	g.Meta       `path:"/voyara/auth/refresh" method:"post" summary:"Refresh access token"`
	RefreshToken string `json:"refreshToken" v:"required"`
}

type ForgotPasswordReq struct {
	g.Meta `path:"/voyara/auth/forgot-password" method:"post" summary:"Send password reset code"`
	Email  string `json:"email" v:"required|email"`
}

type ResetPasswordReq struct {
	g.Meta   `path:"/voyara/auth/reset-password" method:"post" summary:"Reset password with code"`
	Email    string `json:"email" v:"required|email"`
	Code     string `json:"code" v:"required|length:6,6"`
	Password string `json:"password" v:"required|length:8,64"`
}

type ChangePasswordReq struct {
	g.Meta      `path:"/voyara/auth/change-password" method:"post" summary:"Change password"`
	OldPassword string `json:"oldPassword" v:"required"`
	NewPassword string `json:"newPassword" v:"required|length:8,64"`
}

type AuthRes struct {
	Token        string   `json:"token"`
	RefreshToken string   `json:"refreshToken,omitempty"`
	User         UserInfo `json:"user"`
}

type RefreshTokenRes struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type UserInfo struct {
	ID              int    `json:"id"`
	Email           string `json:"email"`
	Name            string `json:"name"`
	Role            string `json:"role"`
	EmailVerified   bool   `json:"emailVerified"`
	PreferredLang   string `json:"preferredLang"`
}

type MessageRes struct {
	Message string `json:"message"`
}

type GetCSRFTokenReq struct {
	g.Meta `path:"/voyara/auth/csrf-token" method:"get" summary:"Get CSRF token"`
}

type GetCSRFTokenRes struct {
	Token string `json:"token"`
}

type GetCurrentUserReq struct {
	g.Meta `path:"/voyara/auth/me" method:"get" summary:"Get current user"`
}

type GetCurrentUserRes struct {
	User UserInfo `json:"user"`
}
