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
	"github.com/ArmanAvanesyan/authsentinel/internal/agent/service"
	"github.com/ArmanAvanesyan/authsentinel/internal/store/redis"
	"github.com/ArmanAvanesyan/authsentinel/pkg/cookie"
	"github.com/ArmanAvanesyan/authsentinel/pkg/token"
)

func main() {
	logger := log.New(os.Stdout, "[authsentinel-agent] ", log.LstdFlags|log.LUTC)
	logger.Println("starting AuthSentinel Agent")

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("config load: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		logger.Fatalf("config invalid: %v", err)
	}

	ctx := context.Background()
	layout := cfg.KeyLayout()
	store, err := redis.New(ctx, cfg.RedisURL, layout)
	if err != nil {
		logger.Fatalf("redis: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			logger.Printf("redis close: %v", err)
		}
	}()

	cookieManager := cookie.NewSignedManager(cfg.CookieSigningSecret)
	jwks := token.NewHTTPJWKSSource(5 * time.Minute)

	svc, err := service.New(cfg, store.SessionStore(), store.PKCEStore(), store.RefreshLockStore(), cookieManager, jwks)
	if err != nil {
		logger.Fatalf("service: %v", err)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: httpserver.New(svc, cfg).Handler(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("http server error: %v", err)
		}
	}()

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-sigCtx.Done()
	logger.Println("shutting down AuthSentinel Agent")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Printf("graceful shutdown failed: %v", err)
	}
}
