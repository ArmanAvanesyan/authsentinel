package sdk

import (
	"net/http"

	"github.com/ArmanAvanesyan/authsentinel/pkg/token"
)

// IdentityFromHTTPRequest extracts an identity principal from an HTTP request.
// TODO: implement header/cookie/token extraction logic.
func IdentityFromHTTPRequest(r *http.Request) (*token.Principal, error) {
	return nil, nil
}
