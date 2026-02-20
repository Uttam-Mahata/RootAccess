package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/badoux/checkmail"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/config"
	mail "github.com/go-mail/mail/v2"
)

type EmailService struct {
	config *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{
		config: cfg,
	}
}

// ValidateEmail checks if email format is valid and domain exists
func (s *EmailService) ValidateEmail(email string) error {
	// Check email format
	if err := checkmail.ValidateFormat(email); err != nil {
		return fmt.Errorf("invalid email format: %w", err)
	}

	// Skipping ValidateHost because port 25 is often blocked in production environments
	// checkmail.ValidateHost(email) would try to dial the MX server

	return nil
}

// GenerateVerificationToken creates a random token for email verification
func (s *EmailService) GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// SendVerificationEmail sends an email with verification link
func (s *EmailService) SendVerificationEmail(toEmail, username, token string) error {
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", s.config.FrontendURL, token)

	subject := "Verify Your Email - RootAccess CTF"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: 'Space Grotesk', Arial, sans-serif; background-color: #0f172a; color: #e2e8f0; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #dc2626 0%%, #991b1b 100%%); padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .header h1 { color: white; margin: 0; font-size: 28px; }
        .content { background-color: #1e293b; padding: 40px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; background: linear-gradient(135deg, #dc2626 0%%, #991b1b 100%%); color: white; text-decoration: none; padding: 15px 40px; border-radius: 8px; font-weight: bold; margin: 20px 0; }
        .button:hover { background: linear-gradient(135deg, #991b1b 0%%, #7f1d1d 100%%); }
        .footer { text-align: center; margin-top: 30px; color: #64748b; font-size: 14px; }
        .code { background-color: #0f172a; padding: 15px; border-radius: 5px; font-family: monospace; color: #f87171; text-align: center; font-size: 18px; letter-spacing: 2px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 RootAccess CTF</h1>
        </div>
        <div class="content">
            <h2 style="color: #f87171;">Welcome to RootAccess, %s!</h2>
            <p>Thank you for registering. To complete your registration and start hacking challenges, please verify your email address.</p>
            <p style="text-align: center;">
                <a href="%s" class="button">Verify Email Address</a>
            </p>
            <p style="color: #94a3b8; font-size: 14px;">Or copy and paste this link into your browser:</p>
            <div class="code">%s</div>
            <p style="color: #94a3b8; font-size: 14px; margin-top: 30px;">This verification link will expire in 24 hours.</p>
            <p style="color: #94a3b8; font-size: 14px;">If you didn't create an account, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>© 2026 RootAccess CTF Platform. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
	`, username, verificationURL, verificationURL)

	return s.sendEmail(toEmail, subject, body)
}

// SendPasswordResetEmail sends an email with password reset link
func (s *EmailService) SendPasswordResetEmail(toEmail, username, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.config.FrontendURL, token)

	subject := "Reset Your Password - RootAccess CTF"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: 'Space Grotesk', Arial, sans-serif; background-color: #0f172a; color: #e2e8f0; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #dc2626 0%%, #991b1b 100%%); padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .header h1 { color: white; margin: 0; font-size: 28px; }
        .content { background-color: #1e293b; padding: 40px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; background: linear-gradient(135deg, #dc2626 0%%, #991b1b 100%%); color: white; text-decoration: none; padding: 15px 40px; border-radius: 8px; font-weight: bold; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; color: #64748b; font-size: 14px; }
        .code { background-color: #0f172a; padding: 15px; border-radius: 5px; font-family: monospace; color: #f87171; text-align: center; font-size: 18px; letter-spacing: 2px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 RootAccess CTF</h1>
        </div>
        <div class="content">
            <h2 style="color: #f87171;">Password Reset Request</h2>
            <p>Hi %s,</p>
            <p>We received a request to reset your password. Click the button below to create a new password:</p>
            <p style="text-align: center;">
                <a href="%s" class="button">Reset Password</a>
            </p>
            <p style="color: #94a3b8; font-size: 14px;">Or copy and paste this link into your browser:</p>
            <div class="code">%s</div>
            <p style="color: #94a3b8; font-size: 14px; margin-top: 30px;">This link will expire in 1 hour.</p>
            <p style="color: #94a3b8; font-size: 14px;">If you didn't request a password reset, please ignore this email or contact support if you have concerns.</p>
        </div>
        <div class="footer">
            <p>© 2026 RootAccess CTF Platform. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
	`, username, resetURL, resetURL)

	return s.sendEmail(toEmail, subject, body)
}

// sendEmail is a helper function to send emails via SMTP
func (s *EmailService) sendEmail(to, subject, body string) error {
	m := mail.NewMessage()
	m.SetHeader("From", s.config.SMTPFrom)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := mail.NewDialer(s.config.SMTPHost, s.config.SMTPPort, s.config.SMTPUser, s.config.SMTPPass)

	// Try to send the email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// VerifyEmailSMTP performs an SMTP verification (checks if mailbox exists)
// Note: Many servers don't support this anymore for anti-spam reasons
func (s *EmailService) VerifyEmailSMTP(email string) error {
	err := checkmail.ValidateHost(email)
	if err != nil {
		return fmt.Errorf("email server validation failed: %w", err)
	}

	// Note: SMTP VRFY command is disabled on most modern mail servers
	// for anti-spam and privacy reasons. Host validation above is sufficient.

	return nil
}

// GetVerificationExpiry returns the expiration time for verification tokens
func (s *EmailService) GetVerificationExpiry() time.Time {
	return time.Now().Add(24 * time.Hour)
}

// GetResetPasswordExpiry returns the expiration time for password reset tokens
func (s *EmailService) GetResetPasswordExpiry() time.Time {
	return time.Now().Add(1 * time.Hour)
}

// SendTeamInvitationEmail sends an email with team invitation link
func (s *EmailService) SendTeamInvitationEmail(toEmail, teamName, inviterName, token string) error {
	invitationURL := fmt.Sprintf("%s/team/invitations?token=%s", s.config.FrontendURL, token)

	subject := fmt.Sprintf("You're invited to join team %s - RootAccess CTF", teamName)
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: 'Space Grotesk', Arial, sans-serif; background-color: #0f172a; color: #e2e8f0; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #dc2626 0%%, #991b1b 100%%); padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .header h1 { color: white; margin: 0; font-size: 28px; }
        .content { background-color: #1e293b; padding: 40px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; background: linear-gradient(135deg, #dc2626 0%%, #991b1b 100%%); color: white; text-decoration: none; padding: 15px 40px; border-radius: 8px; font-weight: bold; margin: 20px 0; }
        .button:hover { background: linear-gradient(135deg, #991b1b 0%%, #7f1d1d 100%%); }
        .footer { text-align: center; margin-top: 30px; color: #64748b; font-size: 14px; }
        .team-name { color: #f87171; font-size: 24px; font-weight: bold; }
        .info-box { background-color: #0f172a; padding: 20px; border-radius: 8px; margin: 20px 0; border-left: 4px solid #dc2626; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>👥 Team Invitation</h1>
        </div>
        <div class="content">
            <h2 style="color: #f87171;">You've been invited!</h2>
            <p>Hi there,</p>
            <p><strong>%s</strong> has invited you to join their team on RootAccess CTF:</p>
            <div class="info-box">
                <p class="team-name">%s</p>
            </div>
            <p>Join forces with your team to solve challenges, earn points, and climb the leaderboard together!</p>
            <p style="text-align: center;">
                <a href="%s" class="button">View Invitation</a>
            </p>
            <p style="color: #94a3b8; font-size: 14px; margin-top: 30px;">This invitation will expire in 7 days.</p>
            <p style="color: #94a3b8; font-size: 14px;">If you don't have an account yet, you'll need to register first before you can join the team.</p>
        </div>
        <div class="footer">
            <p>© 2026 RootAccess CTF Platform. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
	`, inviterName, teamName, invitationURL)

	return s.sendEmail(toEmail, subject, body)
}

// GetTeamInvitationExpiry returns the expiration time for team invitation tokens
func (s *EmailService) GetTeamInvitationExpiry() time.Time {
	return time.Now().Add(7 * 24 * time.Hour) // 7 days
}
