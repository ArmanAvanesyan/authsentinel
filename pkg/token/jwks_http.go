package token

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// HTTPJWKSSource fetches JWKS from the OIDC issuer (via discovery) with an in-memory cache.
type HTTPJWKSSource struct {
	client *http.Client
	cache  map[string]jwkSEntry
	mu     sync.RWMutex
	ttl    time.Duration
}

type jwkSEntry struct {
	data []byte
	exp  time.Time
}

// discoveryResponse is the minimal struct for OIDC discovery to get jwks_uri.
type discoveryResponse struct {
	JWKSURI string `json:"jwks_uri"`
}

// NewHTTPJWKSSource creates an HTTP JWKS source with the given cache TTL.
func NewHTTPJWKSSource(cacheTTL time.Duration) *HTTPJWKSSource {
	if cacheTTL <= 0 {
		cacheTTL = 5 * time.Minute
	}
	return &HTTPJWKSSource{
		client: &http.Client{Timeout: 10 * time.Second},
		cache:  make(map[string]jwkSEntry),
		ttl:    cacheTTL,
	}
}

// GetJWKS returns raw JWKS JSON for the given issuer.
// It fetches issuer/.well-known/openid-configuration to get jwks_uri, then fetches the JWKS.
func (s *HTTPJWKSSource) GetJWKS(ctx context.Context, issuer string) ([]byte, error) {
	issuer = strings.TrimSuffix(issuer, "/")
	cacheKey := issuer

	s.mu.RLock()
	if e, ok := s.cache[cacheKey]; ok && time.Now().Before(e.exp) {
		data := e.data
		s.mu.RUnlock()
		return data, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.cache[cacheKey]; ok && time.Now().Before(e.exp) {
		return e.data, nil
	}

	jwksURI, err := s.fetchJWKSURI(ctx, issuer)
	if err != nil {
		return nil, err
	}
	data, err := s.fetchURL(ctx, jwksURI)
	if err != nil {
		return nil, err
	}
	s.cache[cacheKey] = jwkSEntry{data: data, exp: time.Now().Add(s.ttl)}
	return data, nil
}

func (s *HTTPJWKSSource) fetchJWKSURI(ctx context.Context, issuer string) (string, error) {
	url := issuer + "/.well-known/openid-configuration"
	data, err := s.fetchURL(ctx, url)
	if err != nil {
		return "", fmt.Errorf("oidc discovery: %w", err)
	}
	var dr discoveryResponse
	if err := json.Unmarshal(data, &dr); err != nil {
		return "", fmt.Errorf("oidc discovery parse: %w", err)
	}
	if dr.JWKSURI == "" {
		return "", fmt.Errorf("oidc discovery: missing jwks_uri")
	}
	return dr.JWKSURI, nil
}

func (s *HTTPJWKSSource) fetchURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
