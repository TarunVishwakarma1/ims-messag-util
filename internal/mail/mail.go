package mail

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"ims-message-util/internal/config"
	"log/slog"
	"strings"
	"text/template"
	"time"

	"github.com/wneessen/go-mail"
)

//go:embed templates/*.html
var templateFS embed.FS

type BaseData struct {
	Subject        string
	PreviewText    string
	CompanyURL     string
	CompanyName    string
	RecipientEmail string
	SupportURL     string
	Year           int
}

type OTPData struct {
	BaseData
	OTPCode         string
	ValidityMinutes int
}

type AlertData struct {
	BaseData
	ItemName     string
	SKU          string
	CurrentStock int
	DashboardURL string
}

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
	templates  map[string]*template.Template
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
		mail.WithPassword(cfg.Email.Password),
	)

	if err != nil {
		slog.Error("Error in creating mail client", "ERROR", err)
		return nil, fmt.Errorf("failed to create mail client: %w", err)
	}

	tmplRegistry := make(map[string]*template.Template)

	entries, err := templateFS.ReadDir("templates")

	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, entry := range entries {
		if entry.Name() == "base.html" {
			continue
		}

		tmpl, err := template.ParseFS(templateFS, "templates/base.html", "templates/"+entry.Name())

		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", entry.Name(), err)
		}

		key := strings.TrimSuffix(entry.Name(), ".html")
		tmplRegistry[key] = tmpl
	}

	return &Client{
		smtpClient: c,
		fromEmail:  cfg.Email.Username,
		templates:  tmplRegistry,
	}, nil
}

func (c *Client) sendHTML(ctx context.Context, to, subject, templateName string, data any) error {
	tmpl, exists := c.templates[templateName]
	if !exists {
		return fmt.Errorf("email template '%s' not found in registry", templateName)
	}

	m := mail.NewMsg()
	if err := m.From(c.fromEmail); err != nil {
		return err
	}
	if err := m.To(to); err != nil {
		return err
	}
	m.Subject(subject)

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		return fmt.Errorf("failed to execute template '%s': %w", templateName, err)
	}

	m.SetBodyString(mail.TypeTextHTML, buf.String())

	if err := c.smtpClient.DialAndSendWithContext(ctx, m); err != nil {
		return fmt.Errorf("smtp error: %w", err)
	}

	return nil
}

func (c *Client) OTPMail(ctx context.Context, to string, otpCode string) error {
	data := OTPData{
		BaseData:        c.newBaseData(to, "Your Security Code", "Your secure IMS authentication code is inside."),
		OTPCode:         otpCode,
		ValidityMinutes: 10,
	}

	// Pass "otp" so the engine looks up "otp.html"
	return c.sendHTML(ctx, to, data.Subject, "otp", data)
}

func (c *Client) AlertMail(ctx context.Context, to string, itemName string, currentStock int) error {
	data := AlertData{
		BaseData:     c.newBaseData(to, "Low Stock Alert: "+itemName, "Immediate action required for inventory."),
		ItemName:     itemName,
		SKU:          "UNKNOWN", // Or pass this in as a parameter
		CurrentStock: currentStock,
		DashboardURL: "https://ims.tarunvishwakarma.dev/inventory",
	}

	// Pass "alert" so the engine looks up "alert.html"
	return c.sendHTML(ctx, to, data.Subject, "alert", data)
}

func (c *Client) ReminderMail(ctx context.Context, to []string, subject string, body string) error {
	return nil
}

func (c *Client) newBaseData(to, subject, preview string) BaseData {
	return BaseData{
		Subject:        subject,
		PreviewText:    preview,
		CompanyURL:     "https://ims.tarunvishwakarma.dev",
		CompanyName:    "IMS",
		RecipientEmail: to,
		SupportURL:     "mailto:support@tarunvishwakarma.dev",
		Year:           time.Now().Year(),
	}
}
