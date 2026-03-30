package main

import (
	"context"
	"fmt"
	"ims-message-util/internal/config"
	"ims-message-util/internal/router"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()

	appRouter := router.Setup(cfg)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Application.Port),
		Handler:      appRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("🚀 Server listening on", "PORT", cfg.Application.Port)
		if err := srv.ListenAndServe(); err != nil {
			slog.Error("Error while Listen and serve:", "ERROR", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown:", "ERROR", err)
	}

	slog.Info("Server exited properly")

}
