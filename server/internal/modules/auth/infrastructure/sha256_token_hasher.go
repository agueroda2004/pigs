package infrastructure

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type SHA256TokenHasher struct {
	pepper []byte
}

// NewSHA256TokenHasher builds a token hasher backed by HMAC-SHA256.
// It returns a hasher ready to produce deterministic, peppered token hashes.
func NewSHA256TokenHasher(pepper string) *SHA256TokenHasher {
	return &SHA256TokenHasher{pepper: []byte(pepper)}
}

// Hash returns the hexadecimal HMAC-SHA256 digest of the given token.
// It never fails and always produces a 64-character hash.
func (h *SHA256TokenHasher) Hash(token string) (string, error) {
	mac := hmac.New(sha256.New, h.pepper)
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
