package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
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

	return sendSESEmail(cfg, toEmail, subject, body)
}

func sendSESEmail(cfg EmailConfig, to, subject, body string) error {
	ctx := context.Background()

	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		return fmt.Errorf("aws config: %v", err)
	}

	client := ses.NewFromConfig(awsCfg)

	input := &ses.SendEmailInput{
		Source: &cfg.FromEmail,
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Message: &types.Message{
			Subject: &types.Content{Data: &subject},
			Body: &types.Body{
				Text: &types.Content{Data: &body},
			},
		},
	}

	_, err = client.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("ses send email: %v", err)
	}

	return nil
}

func SendPasswordResetEmail(toEmail, code string) error {
	return SendVerificationCode(toEmail, code, "reset_password")
}

// ── Async order notification emails ──

func sendEmailAsync(to, subject, body string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			if os.Getenv("APP_ENV") != "production" {
				fmt.Printf("[EMAIL] To: %s | Subject: %s\n%s\n\n", to, subject, body)
				done <- nil
				return
			}
			cfg := GetEmailConfig()
			done <- sendSESEmail(cfg, to, subject, body)
		}()

		select {
		case err := <-done:
			if err != nil {
				log.Printf("[EMAIL] Failed to send to %s: %v", to, err)
			}
		case <-ctx.Done():
			log.Printf("[EMAIL] Timeout sending to %s", to)
		}
	}()
}

func SendOrderConfirmationEmail(toEmail, orderNo string) {
	subject := "Order Confirmed - Voyara"
	body := fmt.Sprintf("Thank you for your order!\n\nOrder Number: %s\n\nWe'll notify you when it ships.", orderNo)
	sendEmailAsync(toEmail, subject, body)
}

func SendPaymentSuccessEmail(toEmail, orderNo string) {
	subject := "Payment Received - Voyara"
	body := fmt.Sprintf("Your payment for order %s has been received.\n\nYour order is now being processed.", orderNo)
	sendEmailAsync(toEmail, subject, body)
}

func SendShipmentNotificationEmail(toEmail, orderNo, trackingNumber string) {
	subject := "Order Shipped - Voyara"
	body := fmt.Sprintf("Your order %s has been shipped!\n\nTracking Number: %s", orderNo, trackingNumber)
	sendEmailAsync(toEmail, subject, body)
}
