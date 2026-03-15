package email

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
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
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// buildEmail constructs a RFC 2822 MIME email with a plain-text body and an
// optional file attachment.  Pass attachmentPath = "" to send without an
// attachment.
func buildEmail(to, subject, body, attachmentPath string) (string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// ── RFC 2822 headers ──────────────────────────────────────────────────────
	boundary := writer.Boundary()
	headers := fmt.Sprintf(
		"To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%q\r\n\r\n",
		to, subject, boundary,
	)
	buf.WriteString(headers)

	// ── Plain-text body part ──────────────────────────────────────────────────
	bodyHeader := make(textproto.MIMEHeader)
	bodyHeader.Set("Content-Type", "text/plain; charset=UTF-8")
	bodyHeader.Set("Content-Transfer-Encoding", "quoted-printable")
	bodyPart, err := writer.CreatePart(bodyHeader)
	if err != nil {
		return "", fmt.Errorf("create body part: %w", err)
	}
	fmt.Fprint(bodyPart, body)

	// ── Attachment part (optional) ────────────────────────────────────────────
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

	// Gmail API expects URL-safe base64 (no padding breaks the API)
	return base64.URLEncoding.EncodeToString(buf.Bytes()), nil
}

// sendEmail sends an email via the Gmail API.
func sendEmail(srv *gmail.Service, to, subject, body, attachmentPath string) error {
	rawMsg, err := buildEmail(to, subject, body, attachmentPath)
	if err != nil {
		return fmt.Errorf("build email: %w", err)
	}

	msg := &gmail.Message{Raw: rawMsg}
	_, err = srv.Users.Messages.Send("me", msg).Do()
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	return nil
}

func SendEmail() {
	ctx := context.Background()
	b, err := os.ReadFile("ims-creds.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope, gmail.GmailModifyScope, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	// ── Send an email with a body and an attachment ───────────────────────────
	to := "tarunvishwakarma81@gmail.com"
	subject := "Hello from IMS!"
	body := "Hi,\n\nThis is a test email sent via the Gmail API.\n\nRegards,\nIMS"
	attachmentPath := ""

	if err := sendEmail(srv, to, subject, body, attachmentPath); err != nil {
		log.Fatalf("Unable to send email: %v", err)
	}
	fmt.Println("Email sent successfully!")
}
