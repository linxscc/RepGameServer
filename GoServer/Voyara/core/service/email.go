package service

import (
	"fmt"
	"os"
)

type EmailConfig struct {
	FromEmail string
	Region    string
}

func GetEmailConfig() EmailConfig {
	return EmailConfig{
		FromEmail: envOrDefault("AWS_SES_FROM_EMAIL", "noreply@voyara.com"),
		Region:    envOrDefault("AWS_REGION", "ap-southeast-2"),
	}
}

func SendVerificationCode(toEmail, code, purpose string) error {
	cfg := GetEmailConfig()

	var subject, body string
	switch purpose {
	case "register":
		subject = "Verify your email address"
		body = fmt.Sprintf("Your verification code is: %s\n\nThis code expires in 5 minutes.", code)
	case "reset_password":
		subject = "Reset your password"
		body = fmt.Sprintf("Your password reset code is: %s\n\nThis code expires in 5 minutes.", code)
	default:
		subject = "Your verification code"
		body = fmt.Sprintf("Your verification code is: %s", code)
	}

	if os.Getenv("APP_ENV") != "production" {
		fmt.Printf("[EMAIL] To: %s | Subject: %s | Code: %s\n", toEmail, subject, code)
		return nil
	}

	return sendSESEmail(cfg.FromEmail, toEmail, subject, body)
}

func sendSESEmail(from, to, subject, body string) error {
	fmt.Printf("[SES] From: %s To: %s Subject: %s Body: %s\n", from, to, subject, body)
	return nil
}

func SendPasswordResetEmail(toEmail, code string) error {
	return SendVerificationCode(toEmail, code, "reset_password")
}
