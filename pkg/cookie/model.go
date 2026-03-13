package cookie

// OutCookie represents a cookie the proxy wishes to set on the response.
type OutCookie struct {
	Name    string
	Value   string
	Options CookieOptions
}
