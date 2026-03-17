package email

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"time"
)

var pageTemplates map[string]*template.Template

// InitTemplates pre-parses all email templates into memory.
// templatesPath should be the path to the `templates` directory.
func InitTemplates(templatesPath string) error {
	pageTemplates = make(map[string]*template.Template)

	layoutFiles := []string{
		filepath.Join(templatesPath, "html", "base.html"),
		filepath.Join(templatesPath, "html", "components", "styles.html"),
		filepath.Join(templatesPath, "html", "components", "header.html"),
		filepath.Join(templatesPath, "html", "components", "footer.html"),
	}

	pages := []string{"otp.html", "alert.html", "promotion.html", "newsletter.html"}

	for _, page := range pages {
		pageFile := filepath.Join(templatesPath, "html", page)
		files := append([]string{}, layoutFiles...) // Copy the slice securely
		files = append(files, pageFile)

		t, err := template.ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", page, err)
		}

		pageTemplates[page] = t
	}

	return nil
}

// TemplateData is a wrapper to inject base context data into the templates.
type TemplateData struct {
	Subject string
	Year    int
	Data    interface{}
}

// RenderTemplate renders the specific HTML template against the layout.
// 'page' should match the file name, e.g., "otp.html".
func RenderTemplate(page string, subject string, specificData interface{}) (string, error) {
	t, ok := pageTemplates[page]
	if !ok {
		return "", fmt.Errorf("template %s not found in cache", page)
	}

	data := TemplateData{
		Subject: subject,
		Year:    time.Now().Year(),
		Data:    specificData,
	}

	var buf bytes.Buffer
	// "base.html" defines `{{define "base"}}` layout which brings in "content"
	if err := t.ExecuteTemplate(&buf, "base", data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", page, err)
	}

	return buf.String(), nil
}
