package infrastructure

import (
	"crypto/rand"
	"encoding/base64"
)

type RandomTokenGenerator struct{}

// NewRandomTokenGenerator builds a generator of cryptographically random tokens.
// It returns a generator ready to produce raw refresh tokens.
func NewRandomTokenGenerator() *RandomTokenGenerator {
	return &RandomTokenGenerator{}
}

// Generate returns a 32-byte random token encoded as unpadded base64url.
// It propagates any error from the cryptographic random source.
func (g *RandomTokenGenerator) Generate() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
