package krakend

import (
	"net/http"

	"github.com/ArmanAvanesyan/authsentinel/pkg/cookie"
	"github.com/ArmanAvanesyan/authsentinel/pkg/proxy"
)

// Handler returns an http.Handler that runs the proxy Engine for the KrakenD
// endpoint/auth middleware bridge. Gateway request is translated to proxy.Request;
// Engine.Handle is called; the shared principal decision output (UpstreamHeaders:
// X-User-Id, X-Roles, Authorization, etc.) is set on the response and the request
// is handed off to next or proxied to upstreamURL.
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
		// Allow: set principal decision output as headers for KrakenD backends
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
