package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArmanAvanesyan/authsentinel/internal/proxy"
	"github.com/ArmanAvanesyan/authsentinel/internal/proxy/config"
	"github.com/ArmanAvanesyan/authsentinel/internal/proxy/httpserver"
	"github.com/ArmanAvanesyan/authsentinel/pkg/policy"
	"github.com/ArmanAvanesyan/authsentinel/pkg/observability"
	"github.com/ArmanAvanesyan/authsentinel/pkg/plugindiscovery"
	"github.com/ArmanAvanesyan/authsentinel/pkg/pluginapi"
	"github.com/ArmanAvanesyan/authsentinel/pkg/pluginregistry"
	pkgproxy "github.com/ArmanAvanesyan/authsentinel/pkg/proxy"
	"github.com/ArmanAvanesyan/authsentinel/pkg/plugins/builtin"
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
	if err := (&builtin.Registrar{}).RegisterBuiltins(context.Background(), reg); err != nil {
		logger.Fatalf("register built-in plugins: %v", err)
	}
	if cfg.PluginsManifestDir != "" {
		ctx := context.Background()
		if err := plugindiscovery.DiscoverFromDir(ctx, reg, cfg.PluginsManifestDir, nil); err != nil {
			logger.Printf("plugin discovery: %v (continuing without manifest plugins)", err)
		} else if err := reg.BuildDependencyGraph(); err != nil {
			logger.Printf("plugin dependency graph: %v (continuing)", err)
		}
	}

	metrics, metricsHandler := observability.NewPrometheusMetrics(nil)
	tracer := observability.NewOTLPTracerFromEnv()
	policyEngine, err := buildPolicyEngine(cfg)
	if err != nil {
		logger.Fatalf("policy engine: %v", err)
	}
	pipelinePlugins, err := buildPipelinePlugins(cfg, reg)
	if err != nil {
		logger.Fatalf("pipeline plugins: %v", err)
	}
	handler := httpserver.New(cfg, client, policyEngine, pipelinePlugins, reg, metrics, metricsHandler, tracer).Handler()

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

func buildPolicyEngine(cfg *config.Config) (policy.Engine, error) {
	fallback := policy.FallbackConfig{Allow: cfg.PolicyFallbackAllow != nil && *cfg.PolicyFallbackAllow}

	switch cfg.PolicyEngine {
	case config.PolicyEngineWASM:
		if cfg.PolicyBundlePath == "" {
			return policy.NewWASMRuntime(fallback), nil
		}
		loader := policy.NewBundleLoader(fallback)
		return loader.LoadBundle(cfg.PolicyBundlePath)
	case config.PolicyEngineRego:
		eng := policy.NewRegoEngine(fallback)
		if cfg.PolicyBundlePath != "" {
			if err := eng.Load(cfg.PolicyBundlePath); err != nil {
				return nil, err
			}
		}
		return eng, nil
	default:
		// Should not happen due to ApplyDefaults+Validate, but keep a safe fallback.
		return policy.NewWASMRuntime(fallback), nil
	}
}

func buildPipelinePlugins(cfg *config.Config, reg *pluginregistry.Registry) ([]pkgproxy.PipelinePlugin, error) {
	if reg == nil || len(cfg.PipelinePlugins) == 0 {
		return nil, nil
	}

	ctx := context.Background()
	out := make([]pkgproxy.PipelinePlugin, 0, len(cfg.PipelinePlugins))

	for _, entry := range cfg.PipelinePlugins {
		if entry.ID == "" {
			continue
		}
		regEntry, ok := reg.RegistrationFor(pluginapi.PluginID(entry.ID))
		if !ok || regEntry == nil {
			return nil, fmt.Errorf("pipeline plugin %q not registered", entry.ID)
		}

		p, err := regEntry.Factory(ctx, regEntry.Descriptor)
		if err != nil {
			return nil, fmt.Errorf("pipeline plugin %q factory: %w", entry.ID, err)
		}
		if p == nil {
			return nil, fmt.Errorf("pipeline plugin %q factory returned nil", entry.ID)
		}

		if cp, ok := p.(pluginapi.ConfigurablePlugin); ok {
			if err := cp.Configure(ctx, entry.Raw); err != nil {
				return nil, fmt.Errorf("pipeline plugin %q configure: %w", entry.ID, err)
			}
		}
		if sp, ok := p.(pluginapi.StartablePlugin); ok {
			if err := sp.Start(ctx); err != nil {
				return nil, fmt.Errorf("pipeline plugin %q start: %w", entry.ID, err)
			}
		}

		pl, ok := p.(pluginapi.PipelinePlugin)
		if !ok {
			return nil, fmt.Errorf("pipeline plugin %q is not a PipelinePlugin", entry.ID)
		}

		out = append(out, pl)
	}

	return out, nil
}
