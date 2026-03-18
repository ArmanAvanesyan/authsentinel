package traefik

import (
	"net/http"
	"strings"

	"github.com/ArmanAvanesyan/authsentinel/pkg/cookie"
	"github.com/ArmanAvanesyan/authsentinel/pkg/proxy"
)

// Handler returns an http.Handler that runs the proxy Engine and maps the response
// for Traefik: request context is translated to proxy.Request via proxy.RequestFromHTTP;
// on allow, upstream headers (and optional authResponseHeaders filter) are set and next
// is invoked; on deny, status code and body are written (header and deny response mapping).
func Handler(engine proxy.Engine, next http.Handler, upstreamURL string, authResponseHeaders []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := proxy.RequestFromHTTP(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp, err := engine.Handle(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !resp.Allow {
			proxy.WriteResponse(w, resp)
			return
		}
		// Allow: map upstream headers to response (Traefik forwards these to the backend)
		headersToSet := filterHeaders(resp.UpstreamHeaders, authResponseHeaders)
		for k, v := range headersToSet {
			w.Header().Set(k, v)
		}
		for _, c := range resp.SetCookies {
			cookie.WriteOutCookie(w, c)
		}
		if next != nil {
			next.ServeHTTP(w, r)
			return
		}
		if upstreamURL != "" {
			_ = proxy.ProxyToUpstream(r.Context(), w, r, upstreamURL, resp.UpstreamHeaders, req.Body)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
}

// filterHeaders returns headers to forward. If allowList is empty, forward X-* and Authorization.
func filterHeaders(headers map[string]string, allowList []string) map[string]string {
	if len(allowList) == 0 {
		out := make(map[string]string)
		for k, v := range headers {
			lower := strings.ToLower(k)
			if strings.HasPrefix(lower, "x-") || lower == "authorization" {
				out[k] = v
			}
		}
		return out
	}
	out := make(map[string]string)
	allowed := make(map[string]bool)
	for _, h := range allowList {
		allowed[strings.ToLower(h)] = true
	}
	for k, v := range headers {
		if allowed[strings.ToLower(k)] {
			out[k] = v
		}
	}
	return out
}
