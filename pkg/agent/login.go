package agent

// LoginStartRequest holds input for starting the login flow.
type LoginStartRequest struct {
	// TODO: add fields (redirect URI, scopes, etc.).
}

// LoginStartResponse holds the redirect URL and cookies for login start.
type LoginStartResponse struct {
	// TODO: add fields (redirect URL, cookies, etc.).
}

// LoginEndRequest holds callback params for login end.
type LoginEndRequest struct {
	// TODO: add fields (callback params, cookies, etc.).
}

// LoginEndResponse holds session cookies and redirect target after login.
type LoginEndResponse struct {
	// TODO: add fields (session cookies, redirect target, etc.).
}
