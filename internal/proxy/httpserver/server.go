package httpserver

import "net/http"

// Server is a placeholder HTTP server wrapper for the Proxy app.
type Server struct {
	mux *http.ServeMux
}

// New constructs a new Server instance.
func New() *Server {
	return &Server{
		mux: http.NewServeMux(),
	}
}

// Handler returns the HTTP handler for this server.
func (s *Server) Handler() http.Handler {
	return s.mux
}
