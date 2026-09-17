package infrastructure

import (
	"strings"
	"testing"
)

func TestRandomTokenGeneratorGenerate(t *testing.T) {
	generator := NewRandomTokenGenerator()

	t.Run("produces a base64url token of expected length", func(t *testing.T) {
		token, err := generator.Generate()

		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if len(token) != 43 {
			t.Fatalf("unexpected length: got %d, want 43", len(token))
		}
		if strings.Contains(token, "=") {
			t.Fatalf("token should not contain padding: %q", token)
		}
	})

	t.Run("produces unique tokens", func(t *testing.T) {
		first, err := generator.Generate()
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		second, err := generator.Generate()
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if first == second {
			t.Fatalf("expected unique tokens, got %q twice", first)
		}
	})
}

func TestSHA256TokenHasherHash(t *testing.T) {
	hasher := NewSHA256TokenHasher()

	hash, err := hasher.Hash("token")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if len(hash) != 64 {
		t.Fatalf("unexpected length: got %d, want 64", len(hash))
	}

	again, err := hasher.Hash("token")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if hash != again {
		t.Fatalf("expected deterministic hash, got %q and %q", hash, again)
	}
}
