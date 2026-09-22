package infrastructure

import (
	"crypto/sha256"
	"encoding/hex"
)

type SHA256TokenHasher struct{}

// NewSHA256TokenHasher builds a token hasher backed by SHA-256.
// It returns a hasher ready to produce deterministic token hashes.
func NewSHA256TokenHasher() *SHA256TokenHasher {
	return &SHA256TokenHasher{}
}

// Hash returns the hexadecimal SHA-256 digest of the given token.
// It never fails and always produces a 64-character hash.
func (h *SHA256TokenHasher) Hash(token string) (string, error) {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:]), nil
}
