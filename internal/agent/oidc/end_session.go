package oidc

import (
	"context"
	"net/url"
	"strings"
)

// EndSessionURL returns the URL to redirect the user to for logout (end_session_endpoint).
func (c *Client) EndSessionURL(ctx context.Context, idTokenHint, postLogoutRedirectURI string) (string, error) {
	disc, err := c.Discovery(ctx)
	if err != nil {
		return "", err
	}
	if disc.EndSessionEndpoint == "" {
		return "", nil
	}
	params := url.Values{}
	if idTokenHint != "" {
		params.Set("id_token_hint", idTokenHint)
	}
	if postLogoutRedirectURI != "" {
		params.Set("post_logout_redirect_uri", postLogoutRedirectURI)
	}
	q := params.Encode()
	if q != "" {
		return strings.TrimSuffix(disc.EndSessionEndpoint, "/") + "?" + q, nil
	}
	return disc.EndSessionEndpoint, nil
}
