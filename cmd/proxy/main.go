package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArmanAvanesyan/authsentinel/internal/proxy"
	"github.com/ArmanAvanesyan/authsentinel/internal/proxy/config"
	"github.com/ArmanAvanesyan/authsentinel/internal/proxy/httpserver"
)

func main() {
	logger := log.New(os.Stdout, "[authsentinel-proxy] ", log.LstdFlags|log.LUTC)
	logger.Println("starting AuthSentinel Proxy")

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("config load: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		logger.Fatalf("config invalid: %v", err)
	}

	client := proxy.NewAgentClient(cfg.AgentURL, cfg.CookieName)

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: httpserver.New(cfg, client).Handler(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("http server error: %v", err)
		}
	}()

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-sigCtx.Done()
	logger.Println("shutting down AuthSentinel Proxy")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Printf("graceful shutdown failed: %v", err)
	}
}
