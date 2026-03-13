package proxy

import (
	"context"
	"net/http"
)

// Middleware returns an HTTP middleware that delegates to the proxy Engine.
// TODO: implement; convert http.Request to Request, call Engine.Handle, write Response.
func Middleware(e Engine) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = e
			_ = context.Background()
			next.ServeHTTP(w, r)
		})
	}
}
