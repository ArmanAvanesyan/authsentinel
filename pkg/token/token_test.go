package token

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNormalizeClaims(t *testing.T) {
	// nil -> nil
	if got := NormalizeClaims(nil); got != nil {
		t.Errorf("NormalizeClaims(nil) = %v, want nil", got)
	}
	// roles from realm_access.roles
	in := map[string]any{
		"sub": "u1",
		"realm_access": map[string]any{"roles": []interface{}{"a", "b"}},
	}
	out := NormalizeClaims(in)
	if out == nil {
		t.Fatal("expected non-nil map")
	}
	roles, ok := out["roles"].([]string)
	if !ok || len(roles) != 2 || roles[0] != "a" || roles[1] != "b" {
		t.Errorf("expected roles [a b], got %#v", out["roles"])
	}
	// roles from top-level roles
	in2 := map[string]any{"roles": []interface{}{"x"}}
	out2 := NormalizeClaims(in2)
	roles2, _ := out2["roles"].([]string)
	if len(roles2) != 1 || roles2[0] != "x" {
		t.Errorf("expected roles [x], got %#v", roles2)
	}
}

func TestKeyFuncFromJWKS_ECDSA(t *testing.T) {
	// Generate a P-256 key and build minimal JWKS
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	pub := &priv.PublicKey
	xBytes := pub.X.Bytes()
	yBytes := pub.Y.Bytes()
	// JWK uses fixed width for x,y; ensure we pad to curve size
	curveSize := (elliptic.P256().Params().BitSize + 7) / 8
	if len(xBytes) < curveSize {
		padded := make([]byte, curveSize)
		copy(padded[curveSize-len(xBytes):], xBytes)
		xBytes = padded
	}
	if len(yBytes) < curveSize {
		padded := make([]byte, curveSize)
		copy(padded[curveSize-len(yBytes):], yBytes)
		yBytes = padded
	}
	jwks := map[string]any{
		"keys": []map[string]any{
			{
				"kid": "ec1",
				"kty": "EC",
				"crv": "P-256",
				"x":   base64.RawURLEncoding.EncodeToString(xBytes),
				"y":   base64.RawURLEncoding.EncodeToString(yBytes),
			},
		},
	}
	jwksData, _ := json.Marshal(jwks)
	keyfunc := keyFuncFromJWKS(jwksData)
	token := &jwt.Token{Header: map[string]any{"kid": "ec1", "alg": "ES256"}}
	key, err := keyfunc(token)
	if err != nil {
		t.Fatalf("keyfunc: %v", err)
	}
	ecPub, ok := key.(*ecdsa.PublicKey)
	if !ok {
		t.Fatalf("expected *ecdsa.PublicKey, got %T", key)
	}
	if ecPub.Curve != elliptic.P256() || ecPub.X.Cmp(pub.X) != 0 || ecPub.Y.Cmp(pub.Y) != 0 {
		t.Error("EC key from JWKS did not match original public key")
	}
}

func TestJWTValidator_ValidateJWT(t *testing.T) {
	// Mock JWKSSource returning RSA JWKS; sign a JWT with matching key
	priv, err := rsaPrivateKeyForTest()
	if err != nil {
		t.Fatalf("rsa key: %v", err)
	}
	jwksData := rsaJWKSForTest(t, priv)
	issuer := "https://issuer.example"
	audience := "my-client"
	mock := &mockJWKSSource{jwks: map[string][]byte{issuer: jwksData}}
	v := NewJWTValidator(mock, issuer, audience)
	raw, err := signTestJWT(priv, issuer, audience, "user-1", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	ctx := context.Background()
	principal, err := v.ValidateJWT(ctx, raw)
	if err != nil {
		t.Fatalf("ValidateJWT: %v", err)
	}
	if principal.Subject != "user-1" {
		t.Errorf("Subject = %q, want user-1", principal.Subject)
	}
}

func TestJWTValidator_IntrospectOpaque(t *testing.T) {
	v := NewJWTValidator(nil, "", "")
	_, err := v.IntrospectOpaque(context.Background(), "opaque")
	if err != ErrIntrospectNotSupported {
		t.Errorf("IntrospectOpaque err = %v, want ErrIntrospectNotSupported", err)
	}
}

func TestPrincipalFields(t *testing.T) {
	now := time.Now()
	p := &Principal{
		Subject:   "user-123",
		Scopes:    []string{"read", "write"},
		Roles:     []string{"admin"},
		Claims:    map[string]any{"foo": "bar"},
		ExpiresAt: now,
	}

	if p.Subject != "user-123" {
		t.Errorf("expected Subject %q, got %q", "user-123", p.Subject)
	}
	if len(p.Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(p.Scopes))
	}
	if len(p.Roles) != 1 || p.Roles[0] != "admin" {
		t.Errorf("unexpected roles: %#v", p.Roles)
	}
	if got := p.Claims["foo"]; got != "bar" {
		t.Errorf("expected claim foo=bar, got %#v", got)
	}
	if !p.ExpiresAt.Equal(now) {
		t.Errorf("expected ExpiresAt %v, got %v", now, p.ExpiresAt)
	}
}

type mockJWKSSource struct {
	jwks map[string][]byte
}

func (m *mockJWKSSource) GetJWKS(ctx context.Context, issuer string) ([]byte, error) {
	if b, ok := m.jwks[issuer]; ok {
		return b, nil
	}
	return nil, nil
}

func rsaPrivateKeyForTest() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

func rsaJWKSForTest(t *testing.T, priv *rsa.PrivateKey) []byte {
	t.Helper()
	n := base64.RawURLEncoding.EncodeToString(priv.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(priv.E)).Bytes())
	jwks := map[string]any{
		"keys": []map[string]any{
			{"kid": "rsa1", "kty": "RSA", "n": n, "e": e},
		},
	}
	b, err := json.Marshal(jwks)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func signTestJWT(priv *rsa.PrivateKey, issuer, audience, sub string, exp time.Time) (string, error) {
	claims := jwt.MapClaims{
		"iss": issuer,
		"aud": audience,
		"sub": sub,
		"exp": exp.Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "rsa1"
	return token.SignedString(priv)
}

func signTestJWTWithNonce(priv *rsa.PrivateKey, issuer, audience, sub, nonce string, exp time.Time) (string, error) {
	claims := jwt.MapClaims{
		"iss": issuer, "aud": audience, "sub": sub, "exp": exp.Unix(),
		"iat": time.Now().Unix(), "nonce": nonce,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "rsa1"
	return token.SignedString(priv)
}

func TestValidateIDToken_Success(t *testing.T) {
	priv, err := rsaPrivateKeyForTest()
	if err != nil {
		t.Fatalf("rsa key: %v", err)
	}
	issuer := "https://issuer.example"
	audience := "my-client"
	jwksData := rsaJWKSForTest(t, priv)
	mock := &mockJWKSSource{jwks: map[string][]byte{issuer: jwksData}}
	raw, err := signTestJWTWithNonce(priv, issuer, audience, "user-1", "nonce-123", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	ctx := context.Background()
	p, err := ValidateIDToken(ctx, raw, mock, issuer, audience, "nonce-123")
	if err != nil {
		t.Fatalf("ValidateIDToken: %v", err)
	}
	if p.Subject != "user-1" {
		t.Errorf("Subject = %q, want user-1", p.Subject)
	}
}

func TestValidateIDToken_WrongIssuer(t *testing.T) {
	priv, err := rsaPrivateKeyForTest()
	if err != nil {
		t.Fatalf("rsa key: %v", err)
	}
	jwksData := rsaJWKSForTest(t, priv)
	mock := &mockJWKSSource{jwks: map[string][]byte{"https://issuer.example": jwksData}}
	raw, err := signTestJWT(priv, "https://issuer.example", "client", "sub", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	ctx := context.Background()
	_, err = ValidateIDToken(ctx, raw, mock, "https://wrong-issuer.example", "client", "")
	if err == nil {
		t.Fatal("expected error for wrong issuer")
	}
}

func TestValidateIDToken_WrongAudience(t *testing.T) {
	priv, err := rsaPrivateKeyForTest()
	if err != nil {
		t.Fatalf("rsa key: %v", err)
	}
	jwksData := rsaJWKSForTest(t, priv)
	mock := &mockJWKSSource{jwks: map[string][]byte{"https://issuer.example": jwksData}}
	raw, err := signTestJWT(priv, "https://issuer.example", "client-a", "sub", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	ctx := context.Background()
	_, err = ValidateIDToken(ctx, raw, mock, "https://issuer.example", "client-b", "")
	if err == nil {
		t.Fatal("expected error for wrong audience")
	}
}

func TestValidateIDToken_Expired(t *testing.T) {
	priv, err := rsaPrivateKeyForTest()
	if err != nil {
		t.Fatalf("rsa key: %v", err)
	}
	jwksData := rsaJWKSForTest(t, priv)
	mock := &mockJWKSSource{jwks: map[string][]byte{"https://issuer.example": jwksData}}
	raw, err := signTestJWT(priv, "https://issuer.example", "client", "sub", time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	ctx := context.Background()
	_, err = ValidateIDToken(ctx, raw, mock, "https://issuer.example", "client", "")
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestValidateIDToken_NonceMismatch(t *testing.T) {
	priv, err := rsaPrivateKeyForTest()
	if err != nil {
		t.Fatalf("rsa key: %v", err)
	}
	jwksData := rsaJWKSForTest(t, priv)
	mock := &mockJWKSSource{jwks: map[string][]byte{"https://issuer.example": jwksData}}
	raw, err := signTestJWTWithNonce(priv, "https://issuer.example", "client", "sub", "correct-nonce", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	ctx := context.Background()
	_, err = ValidateIDToken(ctx, raw, mock, "https://issuer.example", "client", "wrong-nonce")
	if err == nil {
		t.Fatal("expected error for nonce mismatch")
	}
}

func TestValidateIDToken_InvalidSignature(t *testing.T) {
	priv, err := rsaPrivateKeyForTest()
	if err != nil {
		t.Fatalf("rsa key: %v", err)
	}
	// JWKS from a different key
	otherPriv, _ := rsa.GenerateKey(rand.Reader, 2048)
	jwksData := rsaJWKSForTest(t, otherPriv)
	mock := &mockJWKSSource{jwks: map[string][]byte{"https://issuer.example": jwksData}}
	raw, err := signTestJWT(priv, "https://issuer.example", "client", "sub", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	ctx := context.Background()
	_, err = ValidateIDToken(ctx, raw, mock, "https://issuer.example", "client", "")
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
}

func TestValidateIDToken_MissingKidInJWKS(t *testing.T) {
	priv, err := rsaPrivateKeyForTest()
	if err != nil {
		t.Fatalf("rsa key: %v", err)
	}
	// JWKS with different kid than token
	jwks := map[string]any{
		"keys": []map[string]any{
			{"kid": "other-kid", "kty": "RSA",
				"n": base64.RawURLEncoding.EncodeToString(priv.N.Bytes()),
				"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(priv.E)).Bytes())},
		},
	}
	jwksData, _ := json.Marshal(jwks)
	mock := &mockJWKSSource{jwks: map[string][]byte{"https://issuer.example": jwksData}}
	raw, err := signTestJWT(priv, "https://issuer.example", "client", "sub", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	ctx := context.Background()
	_, err = ValidateIDToken(ctx, raw, mock, "https://issuer.example", "client", "")
	if err == nil {
		t.Fatal("expected error when token kid not in JWKS")
	}
}

func TestNormalizeClaims_GroupsPreserved(t *testing.T) {
	in := map[string]any{
		"sub":    "u1",
		"groups": []interface{}{"g1", "g2"},
	}
	out := NormalizeClaims(in)
	if out == nil {
		t.Fatal("expected non-nil map")
	}
	groups, ok := out["groups"].([]interface{})
	if !ok || len(groups) != 2 {
		t.Errorf("expected groups [g1 g2], got %#v", out["groups"])
	}
}

func TestValidateIDToken_PrincipalFromClaims(t *testing.T) {
	priv, err := rsaPrivateKeyForTest()
	if err != nil {
		t.Fatalf("rsa key: %v", err)
	}
	jwksData := rsaJWKSForTest(t, priv)
	mock := &mockJWKSSource{jwks: map[string][]byte{"https://issuer.example": jwksData}}
	raw, err := signTestJWT(priv, "https://issuer.example", "client", "alice", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	ctx := context.Background()
	p, err := ValidateIDToken(ctx, raw, mock, "https://issuer.example", "client", "")
	if err != nil {
		t.Fatalf("ValidateIDToken: %v", err)
	}
	if p.Subject != "alice" {
		t.Errorf("Subject = %q, want alice", p.Subject)
	}
	if p.Claims == nil {
		t.Fatal("expected non-nil Claims")
	}
	if p.Claims["sub"] != "alice" {
		t.Errorf("Claims[sub] = %v", p.Claims["sub"])
	}
}
