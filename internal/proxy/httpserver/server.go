package httpserver

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/ArmanAvanesyan/authsentinel/internal/proxy"
	"github.com/ArmanAvanesyan/authsentinel/internal/proxy/config"
)

// Server is the HTTP server for the Proxy app.
type Server struct {
	mux    *http.ServeMux
	cfg    *config.Config
	client *proxy.AgentClient
}

// New constructs a new Server with the given config and agent client.
func New(cfg *config.Config, client *proxy.AgentClient) *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		cfg:    cfg,
		client: client,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	prefix := s.cfg.ProxyPathPrefix
	if prefix == "" {
		prefix = "/"
	}
	if !strings.HasSuffix(prefix, "/") {
		s.mux.Handle(prefix, s.proxyHandler())
	}
	s.mux.Handle(prefix+"/", s.proxyHandler())
}

func (s *Server) proxyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieVal := ""
		if c, _ := r.Cookie(s.cfg.CookieName); c != nil {
			cookieVal = c.Value
		}
		resolve, err := s.client.Resolve(r.Context(), cookieVal)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		if resolve == nil && s.cfg.RequireAuth {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"errors":[{"message":"unauthorized"}]}`))
			return
		}
		var headers map[string]string
		if resolve != nil {
			headers = proxy.BuildUpstreamHeaders(s.cfg, resolve.AccessToken, resolve.Claims, resolve.TenantContext)
		} else {
			headers = make(map[string]string)
		}
		upstream, _ := url.Parse(s.cfg.UpstreamURL)
		proxy := httputil.NewSingleHostReverseProxy(upstream)
		proxy.Director = func(out *http.Request) {
			out.URL.Scheme = upstream.Scheme
			out.URL.Host = upstream.Host
			out.URL.Path = singleJoiningSlash(upstream.Path, r.URL.Path)
			if r.URL.RawQuery != "" {
				out.URL.RawQuery = r.URL.RawQuery
			}
			out.Host = upstream.Host
			for k, v := range r.Header {
				if strings.EqualFold(k, "Cookie") {
					continue
				}
				out.Header[k] = v
			}
			for k, v := range headers {
				out.Header.Set(k, v)
			}
			out.RequestURI = ""
		}
		proxy.ModifyResponse = func(resp *http.Response) error {
			return nil
		}
		proxy.ServeHTTP(w, r)
	})
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

// Handler returns the HTTP handler.
func (s *Server) Handler() http.Handler {
	return s.mux
}
