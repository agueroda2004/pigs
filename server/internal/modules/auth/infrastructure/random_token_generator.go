package infrastructure

import (
	"crypto/rand"
	"encoding/base64"
)

// + === TYPE ===
type RandomTokenGenerator struct{}

// + === CONSTRUCTOR ===
func NewRandomTokenGenerator() *RandomTokenGenerator {
	return &RandomTokenGenerator{}
}

// + === METHODS ===
func (g *RandomTokenGenerator) Generate() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
