package cookie

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

var ErrInvalidSignature = errors.New("cookie: invalid signature")

const signedSep = "."

// SignedManager implements Manager with HMAC-SHA256 signed values (session ID only).
// Only parties with the same secret can create or validate cookies.
type SignedManager struct {
	secret []byte
}

// NewSignedManager returns a Manager that signs and verifies cookie values with the given secret.
func NewSignedManager(secret string) *SignedManager {
	return &SignedManager{secret: []byte(secret)}
}

// Encode signs v (expected to be a string session ID) and returns a cookie-safe value.
func (m *SignedManager) Encode(v any) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", nil
	}
	payload := base64.URLEncoding.EncodeToString([]byte(s))
	sig := m.sign([]byte(payload))
	return payload + signedSep + base64.URLEncoding.EncodeToString(sig), nil
}

// Decode verifies the signature and decodes the session ID into dst (must be *string).
// Returns ErrInvalidSignature when the value is malformed or the signature does not match.
func (m *SignedManager) Decode(raw string, dst any) error {
	parts := strings.SplitN(raw, signedSep, 2)
	if len(parts) != 2 {
		return ErrInvalidSignature
	}
	payload, sigB64 := parts[0], parts[1]
	sig, err := base64.URLEncoding.DecodeString(sigB64)
	if err != nil {
		return ErrInvalidSignature
	}
	expected := m.sign([]byte(payload))
	if !hmac.Equal(sig, expected) {
		return ErrInvalidSignature
	}
	dec, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return ErrInvalidSignature
	}
	if ptr, ok := dst.(*string); ok {
		*ptr = string(dec)
	}
	return nil
}

func (m *SignedManager) sign(data []byte) []byte {
	h := hmac.New(sha256.New, m.secret)
	h.Write(data)
	return h.Sum(nil)
}

// Set writes a cookie with the given name, value, and options.
func (m *SignedManager) Set(w http.ResponseWriter, name string, value string, opts CookieOptions) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     opts.Path,
		Domain:   opts.Domain,
		MaxAge:   opts.MaxAge,
		Secure:   opts.Secure,
		HttpOnly: opts.HTTPOnly,
		SameSite: opts.SameSite,
	}
	if opts.Path == "" {
		c.Path = "/"
	}
	http.SetCookie(w, c)
}

// Clear clears the cookie by setting MaxAge=-1.
func (m *SignedManager) Clear(w http.ResponseWriter, name string, opts CookieOptions) {
	c := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     opts.Path,
		Domain:   opts.Domain,
		MaxAge:   -1,
		Secure:   opts.Secure,
		HttpOnly: opts.HTTPOnly,
		SameSite: opts.SameSite,
	}
	if opts.Path == "" {
		c.Path = "/"
	}
	http.SetCookie(w, c)
}
