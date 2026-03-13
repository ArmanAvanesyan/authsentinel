package token

// ParseJWT parses a raw JWT string without verification.
// TODO: implement; used by Validator implementations.
func ParseJWT(raw string) (header, payload, signature []byte, err error) {
	return nil, nil, nil, nil
}
