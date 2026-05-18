package service

import (
	"fmt"

	"github.com/Emerald211/healthconnect/internal/config"
	"github.com/resend/resend-go/v2"
)

type EmailService struct {
	client *resend.Client
	cfg    *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	client := resend.NewClient(cfg.ResendAPIKey)
	return &EmailService{
		client: client,
		cfg:    cfg,
	}
}

func (s *EmailService) GetSubject(purpose string) string {
	switch purpose {
	case "verify_email":
		return "Verify your email"
	case "reset_password":
		return "Reset your password"
	default:
		return "HealthConnect"
	}
}

func (s *EmailService) buildOTPEmail(name, otp, purpose string) string {
	action := "verify your email address"
	if purpose == "reset_password" {
		action = "reset your password"
	}

	return fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #2563eb;">HealthConnect</h2>
			<p>Hi %s,</p>
			<p>Use the OTP below to %s. This code expires in <strong>10 minutes</strong>.</p>
			<div style="background: #f3f4f6; padding: 20px; text-align: center; border-radius: 8px; margin: 24px 0;">
				<h1 style="letter-spacing: 8px; color: #2563eb; font-size: 36px; margin: 0;">%s</h1>
			</div>
			<p style="color: #6b7280; font-size: 14px;">If you did not request this, please ignore this email.</p>
			<p>— The HealthConnect Team</p>
		</div>
	`, name, action, otp)
}

func (s *EmailService) SendOTP(email, name, otp, purpose string) error {
	subject := s.GetSubject(purpose)
	html := s.buildOTPEmail(name, otp, purpose)

	params := &resend.SendEmailRequest{
		From:    s.cfg.EmailFrom,
		To:      []string{email},
		Subject: subject,
		Html:    html,
	}

	_, err := s.client.Emails.Send(params)

	if err != nil {
		return fmt.Errorf("sending email: %w", err)
	}

	return nil

}
