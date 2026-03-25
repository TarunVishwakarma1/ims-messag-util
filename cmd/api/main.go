package main

import (
	"context"
	"encoding/json"
	"fmt"
	"ims-message-util/internal/config"
	"ims-message-util/internal/mail"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type OTPData struct {
	To  string `json:"to"`
	OTP string `json:"otp"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/email", email)
	http.ListenAndServe(":8080", r)
}

func email(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var otpdata OTPData
	data, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Erorr in reading data", "ERROR", err)
	}

	c := config.Load()

	mailer, err := mail.NewClient(c)
	if err != nil {
		slog.Error("Error while creating email client", "ERROR", err)
	}

	json.Unmarshal(data, &otpdata)
	fmt.Println(string(data), otpdata.OTP, otpdata.To)
	mailer.OTPMail(ctx, otpdata.To, otpdata.OTP)
	w.WriteHeader(200)
	w.Write([]byte("Sent"))

}
