package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArmanAvanesyan/authsentinel/internal/agent/config"
	"github.com/ArmanAvanesyan/authsentinel/internal/agent/httpserver"
)

func main() {
	logger := log.New(os.Stdout, "[authsentinel-agent] ", log.LstdFlags|log.LUTC)
	logger.Println("starting AuthSentinel Agent (skeleton)")

	_, _ = config.Load()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: httpserver.New().Handler(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("http server error: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	logger.Println("shutting down AuthSentinel Agent")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Printf("graceful shutdown failed: %v", err)
	}
}
