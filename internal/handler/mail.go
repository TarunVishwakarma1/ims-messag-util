package handler

import (
	"ims-message-util/internal/mail"
	"io"
	"log/slog"
	"net/http"
)

type MailHandler struct {
	client *mail.Client
}

func NewMailHandler(client *mail.Client) *MailHandler {
	return &MailHandler{client: client}
}

func (m *MailHandler) SendOTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read request body", "ERROR", err)
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := m.client.OTPMail(r.Context(), body); err != nil {
		slog.Error("Failed to send OTP mail", "ERROR", err)
		http.Error(w, `{"error": "Failed to send OTP mail"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "OTP sent successfully"}`))
}

func (m *MailHandler) SendAlert(w http.ResponseWriter, r *http.Request) {
	println("Sending Alert")
	// TODO: to be implemented later
}

func (m *MailHandler) SendReminder(w http.ResponseWriter, r *http.Request) {
	println("Sending Reminder")
	// TODO: to be implemented later
}
