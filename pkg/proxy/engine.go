package proxy

import (
	"context"
	"net/http"
	"strings"

	"github.com/ArmanAvanesyan/authsentinel/pkg/observability"
	"github.com/ArmanAvanesyan/authsentinel/pkg/policy"
	"github.com/ArmanAvanesyan/authsentinel/pkg/token"
)

// PrincipalResolver resolves the principal (identity) from a proxy request (e.g. via session cookie or JWT).
type PrincipalResolver interface {
	Resolve(ctx context.Context, req *Request) (*token.Principal, error)
}

// HeaderBuilder builds upstream headers from principal and policy obligations.
// If nil, DefaultEngine uses a default that sets X-User-Id, X-Roles, and obligation-derived headers.
type HeaderBuilder func(principal *token.Principal, obligations map[string]any) map[string]string

// DefaultEngine implements Engine: normalizes request, resolves principal, evaluates policy, maps decision to response.
type DefaultEngine struct {
	Resolver      PrincipalResolver
	Policy        policy.Engine
	UpstreamURL   string
	RequireAuth   bool
	HeaderBuilder HeaderBuilder
	Metrics       observability.Metrics // optional; records auth decisions
}

// Handle implements Engine.Handle.
func (e *DefaultEngine) Handle(ctx context.Context, req *Request) (*Response, error) {
	principal, err := e.Resolver.Resolve(ctx, req)
	if err != nil {
		return &Response{
			Allow:      false,
			StatusCode: http.StatusBadGateway,
			Body:       []byte(`{"errors":[{"message":"auth resolution failed"}]}`),
		}, nil
	}
	if principal == nil && e.RequireAuth {
		return &Response{
			Allow:      false,
			StatusCode: http.StatusUnauthorized,
			Body:       []byte(`{"errors":[{"message":"unauthorized"}]}`),
		}, nil
	}

	input := policy.Input{
		Protocol:         req.Protocol,
		Method:           req.Method,
		Path:             req.Path,
		GraphQLOperation: req.GraphQLOperation,
		GRPCService:      req.GRPCService,
		GRPCMethod:       req.GRPCMethod,
		Principal:        principal,
		Headers:          req.Headers,
	}
	decision, err := e.Policy.Evaluate(ctx, input)
	if err != nil {
		return &Response{
			Allow:      false,
			StatusCode: http.StatusInternalServerError,
			Body:       []byte(`{"errors":[{"message":"policy error"}]}`),
		}, nil
	}
	if decision == nil {
		decision = &policy.Decision{Allow: false, StatusCode: http.StatusServiceUnavailable}
	}

	resp := &Response{
		Allow:      decision.Allow,
		StatusCode: decision.StatusCode,
		Body:       nil,
	}
	if decision.Reason != "" && !decision.Allow {
		resp.Body = []byte(`{"errors":[{"message":"` + escapeJSON(decision.Reason) + `"}]}`)
	}
	if decision.Headers != nil {
		resp.UpstreamHeaders = make(map[string]string)
		for k, v := range decision.Headers {
			resp.UpstreamHeaders[k] = v
		}
	}
	// Merge obligations into headers (e.g. "set_header_X_User" -> "X-User")
	if decision.Obligations != nil {
		if resp.UpstreamHeaders == nil {
			resp.UpstreamHeaders = make(map[string]string)
		}
		for k, v := range decision.Obligations {
			if s, ok := v.(string); ok && strings.HasPrefix(k, "set_header_") {
				headerName := strings.TrimPrefix(k, "set_header_")
				headerName = strings.ReplaceAll(headerName, "_", "-")
				resp.UpstreamHeaders[headerName] = s
			}
		}
	}
	// Build headers from principal when allowed
	if decision.Allow && principal != nil {
		hb := e.HeaderBuilder
		if hb == nil {
			hb = defaultHeaderBuilder
		}
		built := hb(principal, decision.Obligations)
		if resp.UpstreamHeaders == nil {
			resp.UpstreamHeaders = built
		} else {
			for k, v := range built {
				resp.UpstreamHeaders[k] = v
			}
		}
	}
	if e.Metrics != nil {
		e.Metrics.AuthDecision(resp.Allow, resp.StatusCode)
	}
	return resp, nil
}

func defaultHeaderBuilder(principal *token.Principal, obligations map[string]any) map[string]string {
	h := make(map[string]string)
	if principal.AccessToken != "" {
		h["Authorization"] = "Bearer " + principal.AccessToken
	}
	if principal.Subject != "" {
		h["X-User-Id"] = principal.Subject
	}
	if len(principal.Roles) > 0 {
		h["X-Roles"] = strings.Join(principal.Roles, ",")
	}
	if principal.Claims != nil {
		if v, ok := principal.Claims["email"].(string); ok && v != "" {
			h["X-User-Email"] = v
		}
		if v, ok := principal.Claims["name"].(string); ok && v != "" {
			h["X-User-Full-Name"] = v
		}
		if v, ok := principal.Claims["preferred_username"].(string); ok && v != "" {
			h["X-User-Preferred-Username"] = v
		}
	}
	if principal.TenantContext != nil {
		if v, ok := principal.TenantContext["tenant_id"].(string); ok && v != "" {
			h["X-Tenant-Id"] = v
		}
	}
	return h
}

func escapeJSON(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`)
}
