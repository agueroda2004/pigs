package infrastructure

import (
	"crypto/sha256"
	"encoding/hex"
)

// + === TYPE ===
type SHA256TokenHasher struct{}

// + === CONSTRUCTOR ===
func NewSHA256TokenHasher() *SHA256TokenHasher {
	return &SHA256TokenHasher{}
}

// + === METHODS ===
func (h *SHA256TokenHasher) Hash(token string) (string, error) {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:]), nil
}
