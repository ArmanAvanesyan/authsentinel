package graphql

import (
	"net/http"

	"github.com/ArmanAvanesyan/authsentinel/pkg/proxy"
)

// NormalizeHTTPRequest converts an HTTP GraphQL request into a proxy.Request.
// TODO: implement HTTP normalization for GraphQL.
func NormalizeHTTPRequest(r *http.Request) (*proxy.Request, error) {
	return nil, nil
}
