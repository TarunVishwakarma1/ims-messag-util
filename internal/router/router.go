package router

import (
	"ims-message-util/internal/config"
	"ims-message-util/internal/handler"
	"ims-message-util/internal/mail"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Setup(cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	mailClient, err := mail.NewClient(cfg)
	if err != nil {
		slog.Error("Failed to initialize mail client", "error", err)
		os.Exit(1)
	}
	mailHandler := handler.NewMailHandler(mailClient)

	r.Post("/send-otp", mailHandler.SendOTP)
	r.Post("/send-alert", mailHandler.SendAlert)
	r.Post("/send-reminder", mailHandler.SendReminder)

	return r
}
