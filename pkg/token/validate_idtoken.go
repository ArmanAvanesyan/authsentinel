package token

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ValidateIDToken verifies the ID token signature using JWKSSource, then validates exp, iss, aud, and optional nonce.
// Returns a Principal with claims from the token.
func ValidateIDToken(ctx context.Context, raw string, jwksSource JWKSSource, issuer, audience, expectedNonce string) (*Principal, error) {
	if jwksSource == nil {
		return nil, fmt.Errorf("token: JWKSSource is nil")
	}
	jwksData, err := jwksSource.GetJWKS(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("token: get jwks: %w", err)
	}
	keyfunc := keyFuncFromJWKS(jwksData)
	token, err := jwt.Parse(raw, keyfunc, jwt.WithIssuer(issuer), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil {
		return nil, fmt.Errorf("token: parse: %w", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("token: invalid claims")
	}
	// aud: can be string or []string
	if audience != "" {
		audValid := false
		switch v := claims["aud"].(type) {
		case string:
			audValid = v == audience
		case []interface{}:
			for _, a := range v {
				if s, ok := a.(string); ok && s == audience {
					audValid = true
					break
				}
			}
		}
		if !audValid {
			return nil, fmt.Errorf("token: audience mismatch")
		}
	}
	if expectedNonce != "" {
		nonce, _ := claims["nonce"].(string)
		if nonce != expectedNonce {
			return nil, fmt.Errorf("token: nonce mismatch")
		}
	}
	// Build Principal
	sub, _ := claims["sub"].(string)
	exp := time.Time{}
	if v, ok := claims["exp"]; ok {
		switch t := v.(type) {
		case float64:
			exp = time.Unix(int64(t), 0)
		case json.Number:
			n, _ := t.Int64()
			exp = time.Unix(n, 0)
		}
	}
	claimsMap := make(map[string]any)
	for k, v := range claims {
		claimsMap[k] = v
	}
	var roles []string
	if r, ok := claims["realm_access"].(map[string]any); ok {
		if arr, ok := r["roles"].([]interface{}); ok {
			for _, x := range arr {
				if s, ok := x.(string); ok {
					roles = append(roles, s)
				}
			}
		}
	}
	if len(roles) == 0 {
		if r, ok := claims["roles"].([]interface{}); ok {
			for _, x := range r {
				if s, ok := x.(string); ok {
					roles = append(roles, s)
				}
			}
		}
	}
	return &Principal{
		Subject:   sub,
		Roles:     roles,
		Claims:    claimsMap,
		ExpiresAt: exp,
	}, nil
}

// keyFuncFromJWKS returns a jwt.Keyfunc that selects the key by kid from the JWKS JSON.
func keyFuncFromJWKS(jwksData []byte) jwt.Keyfunc {
	var set struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(jwksData, &set); err != nil {
		return func(t *jwt.Token) (interface{}, error) {
			return nil, err
		}
	}
	return func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("token: missing kid in header")
		}
		for _, k := range set.Keys {
			if k.Kid == kid && k.Kty == "RSA" {
				nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
				if err != nil {
					return nil, err
				}
				eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
				if err != nil {
					return nil, err
				}
				n := new(big.Int).SetBytes(nBytes)
				eBig := new(big.Int).SetBytes(eBytes)
				if !eBig.IsInt64() || eBig.Int64() > 1<<31-1 {
					return nil, fmt.Errorf("token: exponent too large")
				}
				return &rsa.PublicKey{N: n, E: int(eBig.Int64())}, nil
			}
		}
		return nil, fmt.Errorf("token: no key found for kid %s", kid)
	}
}
