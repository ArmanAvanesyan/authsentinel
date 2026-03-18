package caddy

import (
	"net/http"

	"github.com/ArmanAvanesyan/authsentinel/pkg/cookie"
	"github.com/ArmanAvanesyan/authsentinel/pkg/proxy"
)

// Handler returns an http.Handler that runs the proxy Engine for each request.
// Gateway request is translated to proxy.Request via proxy.RequestFromHTTP;
// Engine.Handle is called; on allow, upstream headers are set and next is invoked
// (or request is proxied to upstreamURL if next is nil). On deny, the proxy
// response is written (status, headers, body).
//
// This is the shared AuthSentinel runtime integration: all Caddy traffic
// passes through the same proxy/policy/token engine.
func Handler(engine proxy.Engine, next http.Handler, upstreamURL string) http.Handler {
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
		// Allow: add upstream headers and hand off
		for k, v := range resp.UpstreamHeaders {
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
