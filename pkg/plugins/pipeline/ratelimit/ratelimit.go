package ratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ArmanAvanesyan/authsentinel/pkg/pluginapi"
	"github.com/ArmanAvanesyan/authsentinel/pkg/policy"
	"github.com/ArmanAvanesyan/authsentinel/pkg/proxy"
	"github.com/ArmanAvanesyan/authsentinel/pkg/token"
)

// Config matches schemas/plugins/pipeline/ratelimit.schema.json.
type Config struct {
	Name              string `json:"name"`
	RequestsPerMinute int   `json:"requests_per_minute"`
	Burst             int   `json:"burst"`
	KeyStrategy       string `json:"key_strategy"` // ip|principal|header
	HeaderName       string `json:"header_name"`
}

type bucket struct {
	tokens float64
	last   time.Time
}

type Plugin struct {
	mu sync.Mutex

	cfg Config
	// buckets are keyed by derived limit key.
	buckets map[string]*bucket
}

var _ pluginapi.PipelinePlugin = (*Plugin)(nil)
var _ pluginapi.ConfigurablePlugin = (*Plugin)(nil)

func New() *Plugin {
	return &Plugin{buckets: make(map[string]*bucket)}
}

func (p *Plugin) Descriptor() pluginapi.PluginDescriptor {
	return pluginapi.PluginDescriptor{
		ID:             pluginapi.PluginID("pipeline:ratelimit"),
		Kind:           pluginapi.PluginKindPipeline,
		Name:           "RateLimit",
		Description:    "Rate limiting pipeline plugin",
		Version:        "v1",
		Capabilities:   []pluginapi.Capability{"pipeline:ratelimit"},
		ConfigSchemaRef: "plugins/pipeline/ratelimit",
		VersionInfo: pluginapi.VersionInfo{
			APIVersion:        "",
			MinRuntimeVersion: "",
			MaxRuntimeVersion: "",
		},
	}
}

func (p *Plugin) Configure(ctx context.Context, cfg any) error {
	// cfg is expected to be a map[string]any produced by go-config.
	b, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("ratelimit: marshal config: %w", err)
	}
	var out Config
	if err := json.Unmarshal(b, &out); err != nil {
		return fmt.Errorf("ratelimit: decode config: %w", err)
	}
	if out.RequestsPerMinute < 1 {
		return fmt.Errorf("ratelimit: requests_per_minute must be >= 1")
	}
	if out.Burst < 1 {
		return fmt.Errorf("ratelimit: burst must be >= 1")
	}
	out.KeyStrategy = strings.ToLower(strings.TrimSpace(out.KeyStrategy))
	switch out.KeyStrategy {
	case "ip", "principal", "header":
	default:
		return fmt.Errorf("ratelimit: key_strategy must be one of: ip, principal, header")
	}
	if out.KeyStrategy == "header" && strings.TrimSpace(out.HeaderName) == "" {
		return fmt.Errorf("ratelimit: header_name is required when key_strategy=header")
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.cfg = out
	if p.buckets == nil {
		p.buckets = make(map[string]*bucket)
	}
	return nil
}

func (p *Plugin) Health(ctx context.Context) pluginapi.PluginHealth {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cfg.RequestsPerMinute == 0 || p.cfg.Burst == 0 {
		return pluginapi.PluginHealth{
			State:   pluginapi.PluginStateDegraded,
			Message: "not configured",
		}
	}
	return pluginapi.PluginHealth{
		State:   pluginapi.PluginStateHealthy,
		Message: fmt.Sprintf("configured (%s)", p.cfg.KeyStrategy),
		Details: map[string]any{"buckets": len(p.buckets)},
	}
}

func (p *Plugin) Handle(ctx context.Context, req *proxy.Request, principal *token.Principal) (*policy.Decision, error) {
	key := p.deriveKey(req, principal)

	now := time.Now()
	rate := float64(p.cfg.RequestsPerMinute) / 60.0 // tokens per second
	capacity := float64(p.cfg.Burst)

	p.mu.Lock()
	defer p.mu.Unlock()

	b := p.buckets[key]
	if b == nil {
		b = &bucket{tokens: capacity, last: now}
		p.buckets[key] = b
	}

	// Refill tokens based on elapsed time.
	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * rate
		if b.tokens > capacity {
			b.tokens = capacity
		}
	}
	b.last = now

	if b.tokens >= 1 {
		b.tokens -= 1
		return nil, nil
	}

	return &policy.Decision{
		Allow:      false,
		StatusCode: 429,
		Reason:     "rate limit exceeded",
	}, nil
}

func (p *Plugin) deriveKey(req *proxy.Request, principal *token.Principal) string {
	switch p.cfg.KeyStrategy {
	case "principal":
		if principal == nil || strings.TrimSpace(principal.Subject) == "" {
			return "anonymous"
		}
		return principal.Subject
	case "header":
		if req == nil || req.Headers == nil {
			return "anonymous"
		}
		if v := strings.TrimSpace(req.Headers[p.cfg.HeaderName]); v != "" {
			return v
		}
		return "anonymous"
	case "ip":
		// Best-effort: parse X-Forwarded-For first, then common alternatives.
		// proxy.Request stores all headers in lowercase as provided by net/http.
		if req == nil || req.Headers == nil {
			return "anonymous"
		}
		if v := strings.TrimSpace(firstHeader(req.Headers, "X-Forwarded-For")); v != "" {
			if ip := firstIP(v); ip != "" {
				return ip
			}
		}
		if v := strings.TrimSpace(firstHeader(req.Headers, "X-Real-IP")); v != "" {
			if ip := firstIP(v); ip != "" {
				return ip
			}
		}
		if v := strings.TrimSpace(firstHeader(req.Headers, "X-Client-IP")); v != "" {
			if ip := firstIP(v); ip != "" {
				return ip
			}
		}
		return "anonymous"
	default:
		return "anonymous"
	}
}

func firstHeader(m map[string]string, key string) string {
	// net/http lowercases header keys. req.Headers is populated from r.Header iteration,
	// preserving canonical keys as stored in the map. Normalize defensively.
	for k, v := range m {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}

func firstIP(v string) string {
	// X-Forwarded-For may contain a comma-separated list.
	parts := strings.Split(v, ",")
	if len(parts) == 0 {
		return ""
	}
	ip := strings.TrimSpace(parts[0])
	// Validate/canonicalize if possible.
	if parsed := net.ParseIP(ip); parsed != nil {
		return parsed.String()
	}
	return ip
}

