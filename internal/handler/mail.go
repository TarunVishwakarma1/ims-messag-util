package handler

import (
	"ims-message-util/internal/mail"
	"ims-message-util/internal/utils"
	"io"
	"log/slog"
	"net/http"
)

type MailHandler struct {
	client *mail.Client
}

// Struct for WebAPI OTP SEND
type OTP struct {
	To  string `json:"to"`
	OTP string `json:"otp"`
}

type Alert struct {
	To       string `json:"to"`
	ItemName string `json:"itemName"`
	Stock    int    `json:"stock"`
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
	var otpBody OTP
	otpBody, err = utils.ConvertBody[OTP](body)
	if err != nil {
		slog.Error("Failed to parse OTP Body from API, ", "ERROR", err, "body", body)
		http.Error(w, `{"error": "Failed to parse request body"}`, http.StatusBadRequest)
		return
	}

	if err := m.client.OTPMail(r.Context(), otpBody.To, otpBody.OTP); err != nil {
		slog.Error("Failed to send OTP mail", "ERROR", err)
		http.Error(w, `{"error": "Failed to send OTP mail"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "OTP sent successfully"}`))
}

func (m *MailHandler) SendAlert(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read request body", "ERROR", err)
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	var alertBody Alert
	alertBody, err = utils.ConvertBody[Alert](body)
	if err != nil {
		slog.Error("Failed to parse OTP Body from API, ", "ERROR", err, "body", body)
		http.Error(w, `{"error": "Failed to parse request body"}`, http.StatusBadRequest)
		return
	}

	if err = m.client.AlertMail(r.Context(), alertBody.To, alertBody.ItemName, alertBody.Stock); err != nil {
		slog.Error("Failed to send OTP mail", "ERROR", err)
		http.Error(w, `{"error": "Failed to send OTP mail"}`, http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Alert sent successfully"}`))

}

func (m *MailHandler) SendReminder(w http.ResponseWriter, r *http.Request) {
	println("Sending Reminder")
	// TODO: to be implemented later
}
