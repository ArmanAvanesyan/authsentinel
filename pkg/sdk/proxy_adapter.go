package sdk

import (
	"context"
	"net/http"

	"github.com/ArmanAvanesyan/authsentinel/pkg/proxy"
	"github.com/ArmanAvanesyan/authsentinel/pkg/token"
)

// ProxyPrincipalExtractor adapts a proxy.PrincipalResolver to PrincipalExtractor.
// It builds a proxy.Request from the HTTP request without reading the body (so the next handler can still read it).
func NewProxyPrincipalExtractor(resolver proxy.PrincipalResolver) PrincipalExtractor {
	return &proxyPrincipalExtractor{resolver: resolver}
}

type proxyPrincipalExtractor struct {
	resolver proxy.PrincipalResolver
}

func (p *proxyPrincipalExtractor) ExtractPrincipal(ctx context.Context, r *http.Request) (*token.Principal, error) {
	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}
	cookies := make(map[string]string)
	for _, c := range r.Cookies() {
		cookies[c.Name] = c.Value
	}
	protocol := "http"
	if r.TLS != nil {
		protocol = "https"
	}
	path := r.URL.Path
	if r.URL.RawQuery != "" {
		path = r.URL.Path + "?" + r.URL.RawQuery
	}
	req := proxy.NormalizeRequest(protocol, r.Method, path, headers, cookies, nil)
	return p.resolver.Resolve(ctx, req)
}
