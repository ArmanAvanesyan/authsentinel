package proxy

import "github.com/ArmanAvanesyan/authsentinel/pkg/cookie"

// Response is the result of proxy evaluation and policy enforcement.
type Response struct {
	Allow           bool
	UpstreamHeaders map[string]string
	SetCookies      []cookie.OutCookie
	StatusCode      int
	Body            []byte
}
