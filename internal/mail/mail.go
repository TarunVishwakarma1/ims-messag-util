package mail

import (
	"context"
	"fmt"
	"ims-message-util/internal/config"
	"ims-message-util/internal/utils"
	"log/slog"

	"github.com/wneessen/go-mail"
)

type Mail struct {
	To             []string `json:"to"`
	CC             []string `json:"cc"`
	BCC            []string `json:"bcc"`
	Subject        string   `json:"subject"`
	Body           string   `json:"body"`
	AttachmentPath string   `json:"attachment_path"`
}

type Send interface {
	OTPMail(ctx context.Context, to string, otp string) error
	AlertMail(ctx context.Context, to []string, alertMessage string) error
	ReminderMail(ctx context.Context, to []string, subject string, body string) error
}

type Client struct {
	smtpClient *mail.Client
	fromEmail  string
}

// Struct for WebAPI OTP SEND
type OTP struct {
	To  string `json:"to"`
	OTP string `json:"otp"`
}

// Struct for All types of mail
type GeneralMail struct {
	To             []string
	CC             []string
	BCC            []string
	Subject        string
	Body           string
	AttachmentPath string
}

func NewClient(cfg *config.Config) (*Client, error) {

	c, err := mail.NewClient(cfg.Email.SMTPHost,
		mail.WithPort(cfg.Email.SMTPPort),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithSSL(),
		mail.WithUsername(cfg.Email.Username),
		mail.WithPassword(cfg.Email.Password))

	if err != nil {
		slog.Error("Error in creating mail client", "ERROR", err)
		return nil, fmt.Errorf("failed to create mail client: %w", err)
	}

	return &Client{
		smtpClient: c,
		fromEmail:  cfg.Email.Username,
	}, nil
}

func (c *Client) OTPMail(ctx context.Context, body []byte) error {
	otpMail, err := utils.ConvertBody[OTP](body)
	if err != nil {
		slog.Error("Failed to convert body to struct", "ERROR", err)
		return err
	}
	to := otpMail.To
	otpCode := otpMail.OTP

	m := mail.NewMsg()

	if err := m.From(c.fromEmail); err != nil {
		return err
	}
	if err := m.To(to); err != nil {
		return err
	}

	m.Subject("One Time OTP")

	// Format a simple HTML body for the OTP
	htmlBody := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; text-align: center;">
			<h2>Your One-Time Password</h2>
			<p>Please use the following code to complete your action:</p>
			<h1 style="color: #4A90E2; letter-spacing: 5px;">%s</h1>
			<p>This code will expire in 10 minutes.</p>
		</div>
	`, otpCode)

	m.SetBodyString(mail.TypeTextHTML, htmlBody)

	// Send it!
	if err := c.smtpClient.DialAndSendWithContext(ctx, m); err != nil {
		return fmt.Errorf("failed to send OTP to %s: %w", to, err)
	}
	return nil
}

func (c *Client) AlertMail(ctx context.Context, to []string, alertMessage string) error {
	return nil
}

func (c *Client) ReminderMail(ctx context.Context, to []string, subject string, body string) error {
	return nil
}
