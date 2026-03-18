package proxy

import (
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/ArmanAvanesyan/authsentinel/pkg/cookie"
)

// RequestFromHTTP builds a Request from http.Request and normalizes it.
func RequestFromHTTP(r *http.Request) (*Request, error) {
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
	var body []byte
	if r.Body != nil {
		body, _ = io.ReadAll(r.Body)
		_ = r.Body.Close()
	}
	protocol := "http"
	if r.TLS != nil {
		protocol = "https"
	}
	path := r.URL.Path
	if r.URL.RawQuery != "" {
		path = r.URL.Path + "?" + r.URL.RawQuery
	}
	req := NormalizeRequest(protocol, r.Method, path, headers, cookies, body)
	return req, nil
}

// WriteResponse writes the proxy Response to the HTTP response writer (status, SetCookies, body).
// Use for deny/error responses. For Allow, use ProxyToUpstream with resp.UpstreamHeaders.
func WriteResponse(w http.ResponseWriter, resp *Response) {
	for _, c := range resp.SetCookies {
		cookie.WriteOutCookie(w, c)
	}
	if resp.StatusCode == 0 {
		resp.StatusCode = http.StatusOK
	}
	w.WriteHeader(resp.StatusCode)
	if len(resp.Body) > 0 {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(resp.Body)
	}
}

// ProxyToUpstream proxies the request to upstreamURL, adding upstreamHeaders to the outgoing request.
// body is the request body (already read from r); pass the same body used to build Request.
func ProxyToUpstream(ctx context.Context, w http.ResponseWriter, r *http.Request, upstreamURL string, upstreamHeaders map[string]string, body []byte) error {
	upstream, err := url.Parse(upstreamURL)
	if err != nil {
		return err
	}
	path := singleJoiningSlash(upstream.Path, r.URL.Path)
	if path != "" && path[0] != '/' {
		path = "/" + path
	}
	if r.URL.RawQuery != "" {
		path = path + "?" + r.URL.RawQuery
	}
	targetURL := upstream.Scheme + "://" + upstream.Host + path
	outReq, err := http.NewRequestWithContext(ctx, r.Method, targetURL, nil)
	if err != nil {
		return err
	}
	if len(body) > 0 {
		outReq.Body = io.NopCloser(strings.NewReader(string(body)))
		outReq.ContentLength = int64(len(body))
	}
	for k, v := range r.Header {
		if strings.EqualFold(k, "Cookie") {
			continue
		}
		outReq.Header[k] = v
	}
	for k, v := range upstreamHeaders {
		outReq.Header.Set(k, v)
	}
	outReq.Host = upstream.Host
	proxy := httputil.NewSingleHostReverseProxy(upstream)
	proxy.ServeHTTP(w, outReq)
	return nil
}

func singleJoiningSlash(a, b string) string {
	a = strings.TrimSuffix(a, "/")
	b = strings.TrimPrefix(b, "/")
	if a == "" {
		return "/" + b
	}
	if b == "" {
		return a
	}
	return a + "/" + b
}
