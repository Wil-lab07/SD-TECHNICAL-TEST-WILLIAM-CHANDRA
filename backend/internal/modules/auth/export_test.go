package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

// MustParseClaims parses a signed JWT string and returns its Claims.
// Only compiled into test binaries — never part of the production binary.
func MustParseClaims(t *testing.T, tokenStr, secret string) *Claims {
	t.Helper()
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(_ *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("MustParseClaims: %v", err)
	}
	return claims
}
