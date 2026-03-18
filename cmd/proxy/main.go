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
	"github.com/ArmanAvanesyan/authsentinel/pkg/observability"
	"github.com/ArmanAvanesyan/authsentinel/pkg/plugindiscovery"
	"github.com/ArmanAvanesyan/authsentinel/pkg/pluginregistry"
	goconfig "github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/format/json"
	"github.com/ArmanAvanesyan/go-config/source/env"
	"github.com/ArmanAvanesyan/go-config/source/file"
)

func main() {
	logger := log.New(os.Stdout, "[authsentinel-proxy] ", log.LstdFlags|log.LUTC)
	logger.Println("starting AuthSentinel Proxy")

	cfg, err := loadConfig()
	if err != nil {
		logger.Fatalf("config load: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		logger.Fatalf("config invalid: %v", err)
	}

	client := proxy.NewAgentClient(cfg.AgentURL, cfg.CookieName)

	reg := pluginregistry.New()
	if cfg.PluginsManifestDir != "" {
		ctx := context.Background()
		if err := plugindiscovery.DiscoverFromDir(ctx, reg, cfg.PluginsManifestDir, nil); err != nil {
			logger.Printf("plugin discovery: %v (continuing without manifest plugins)", err)
		} else if err := reg.BuildDependencyGraph(); err != nil {
			logger.Printf("plugin dependency graph: %v (continuing)", err)
		}
	}

	metrics, metricsHandler := observability.NewPrometheusMetrics(nil)
	handler := httpserver.New(cfg, client, reg, metrics, metricsHandler).Handler()

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: handler,
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

func loadConfig() (*config.Config, error) {
	ctx := context.Background()
	loader := goconfig.New()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = os.Getenv("PROXY_CONFIG")
	}
	if configPath != "" {
		loader = loader.AddSource(file.New(configPath), json.New())
	}
	loader = loader.AddSource(env.New(""))

	var cfg config.Config
	if err := loader.Load(ctx, &cfg); err != nil {
		return nil, err
	}
	cfg.ApplyDefaults()
	return &cfg, nil
}
