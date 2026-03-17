package email

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type OTPMail struct {
	To  string `json:"to"`
	OTP string `json:"otp"`
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// InitService initializes the Gmail service using local credentials.
func InitService(ctx context.Context) (*gmail.Service, error) {
	b, err := os.ReadFile("ims-creds.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %w", err)
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope, gmail.GmailModifyScope, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	client, err := getClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to get http client: %w", err)
	}

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Gmail client: %w", err)
	}
	return srv, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) (*http.Client, error) {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok), nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}
	return tok, nil
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		slog.Error("Unable to cache oauth token", "error", err, "path", path)
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// buildEmail constructs a RFC 2822 MIME email with a plain-text body and an
// optional file attachment. Pass attachmentPath = "" to send without an attachment.
func buildEmail(to, subject, body, attachmentPath string) (string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// RFC 2822 headers
	boundary := writer.Boundary()
	headers := fmt.Sprintf(
		"To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%q\r\n\r\n",
		to, subject, boundary,
	)
	buf.WriteString(headers)

	// Plain-text body part
	bodyHeader := make(textproto.MIMEHeader)
	bodyHeader.Set("Content-Type", "text/plain; charset=UTF-8")
	bodyHeader.Set("Content-Transfer-Encoding", "quoted-printable")
	bodyPart, err := writer.CreatePart(bodyHeader)
	if err != nil {
		return "", fmt.Errorf("create body part: %w", err)
	}
	fmt.Fprint(bodyPart, body)

	// Attachment part (optional)
	if attachmentPath != "" {
		fileData, err := os.ReadFile(attachmentPath)
		if err != nil {
			return "", fmt.Errorf("read attachment: %w", err)
		}

		fileName := filepath.Base(attachmentPath)
		attachHeader := make(textproto.MIMEHeader)
		attachHeader.Set("Content-Type", "application/octet-stream")
		attachHeader.Set("Content-Transfer-Encoding", "base64")
		attachHeader.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))

		attachPart, err := writer.CreatePart(attachHeader)
		if err != nil {
			return "", fmt.Errorf("create attachment part: %w", err)
		}

		encoded := base64.StdEncoding.EncodeToString(fileData)
		// Wrap encoded data at 76 chars per line (RFC 2045)
		for i := 0; i < len(encoded); i += 76 {
			end := min(i+76, len(encoded))
			fmt.Fprintf(attachPart, "%s\r\n", encoded[i:end])
		}
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close writer: %w", err)
	}

	// Gmail API expects URL-safe base64
	return base64.URLEncoding.EncodeToString(buf.Bytes()), nil
}

// sendEmailWithRetry sends an email via the Gmail API with retries, respecting HTTP context.
func sendEmailWithRetry(ctx context.Context, srv *gmail.Service, to, subject, body, attachmentPath string) error {
	rawMsg, err := buildEmail(to, subject, body, attachmentPath)
	if err != nil {
		return fmt.Errorf("build email: %w", err)
	}

	msg := &gmail.Message{Raw: rawMsg}

	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		// Respect the HTTP request context directly in the Gmail API call
		_, err = srv.Users.Messages.Send("me", msg).Context(ctx).Do()
		if err == nil {
			return nil
		}

		slog.Warn("Failed to send email, retrying", "attempt", i+1, "error", err, "to", to)
		
		select {
		case <-ctx.Done():
			// Don't sleep and retry if the request context is already cancelled/timeout hit
			return fmt.Errorf("context cancelled during retries: %w", ctx.Err())
		case <-time.After(time.Duration(2<<i) * time.Second): // Exponential backoff: 2s, 4s, 8s
		}
	}

	return fmt.Errorf("send message failed after %d attempts: %w", maxRetries, err)
}

// SendEmail parses the payload and sends an OTP email.
func SendEmail(ctx context.Context, srv *gmail.Service, reqBody []byte) error {
	var mailData OTPMail
	if err := json.Unmarshal(reqBody, &mailData); err != nil {
		slog.Error("Invalid email payload", "error", err)
		return fmt.Errorf("invalid payload format")
	}

	to := mailData.To
	if !emailRegex.MatchString(to) {
		return fmt.Errorf("invalid email address format")
	}

	if len(mailData.OTP) < 4 || len(mailData.OTP) > 10 {
		return fmt.Errorf("invalid OTP length")
	}

	subject := "Hello from IMS!"
	emailBody := fmt.Sprintf("Hi,\n\nThis is a test email sent via the Gmail API.\n\nYour OTP is %v\n\nRegards,\nIMS", mailData.OTP)
	attachmentPath := ""

	slog.Info("Sending OTP email", "to", to)

	if err := sendEmailWithRetry(ctx, srv, to, subject, emailBody, attachmentPath); err != nil {
		slog.Error("Failed to send email", "error", err, "to", to)
		return fmt.Errorf("delivery failed")
	}

	slog.Info("Email sent successfully", "to", to)
	return nil
}
