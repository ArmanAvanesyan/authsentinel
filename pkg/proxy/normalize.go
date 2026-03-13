package proxy

// NormalizeRequest builds a proxy Request from raw HTTP/gRPC data.
// TODO: implement full normalization from headers, body, cookies.
func NormalizeRequest(protocol, method, path string, headers, cookies map[string]string, body []byte) *Request {
	return &Request{
		Protocol: protocol,
		Method:   method,
		Path:     path,
		Headers:  headers,
		Cookies:  cookies,
		Body:     body,
	}
}
