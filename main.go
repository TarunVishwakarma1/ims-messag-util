package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"ims-message-util/messangers/utils/email"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"google.golang.org/api/gmail/v1"
)

const (
	CONTENT_TYPE string = "Content-Type"
)

var gmailService *gmail.Service

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Initialize Gmail Service
	var err error
	gmailService, err = email.InitService(context.Background())
	if err != nil {
		slog.Error("Failed to initialize Gmail service", "error", err)
		os.Exit(1)
	}
	slog.Info("Gmail service initialized successfully")

	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context
	r.Use(middleware.Timeout(60 * time.Second))

	// Security Headers Middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/notification/{notificationType}", notification)

	// Apply rate limiting: max 3 requests per 1 minute per IP Address
	r.With(httprate.LimitByIP(3, 1*time.Minute)).Post("/email", email1)

	r.Get("/", events)

	slog.Info("Server starting on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func notification(w http.ResponseWriter, r *http.Request) {
	value := r.PathValue("notificationType")
	w.WriteHeader(200)
	w.Write([]byte(value))
	w.Header().Set(CONTENT_TYPE, "application/json")
}

func email1(w http.ResponseWriter, r *http.Request) {
	// Security: Limit request body to 1KB to prevent slowloris/memory exhaustion attacks
	r.Body = http.MaxBytesReader(w, r.Body, 1024)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Error reading request body", "error", err)
		// If payload is over 1KB, MaxBytesReader returns an error. Send a 413 Payload Too Large
		http.Error(w, "Payload too large or invalid", http.StatusRequestEntityTooLarge)
		return
	}

	// if err := email.SendEmail(r.Context(), gmailService, body); err != nil {
	// 	slog.Error("Failed to send email", "error", err)
	// 	http.Error(w, "Failed to send email: "+err.Error(), http.StatusBadRequest)
	// 	return
	// }

	fmt.Println(string(body))

	value := r.PathValue("notificationType")
	w.WriteHeader(200)
	w.Write([]byte(value))
	w.Header().Set(CONTENT_TYPE, "application/json")
}

func events(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(CONTENT_TYPE, "text/event-stream")

	tokens := []string{"this", "is", "a", "live", "east", "test", "from", "you", "tube"}

	for _, token := range tokens {
		content := fmt.Sprintf("data: %s\n", token)
		w.Write([]byte(content))
		w.(http.Flusher).Flush()
	}
}
