package infrastructure

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	userdomain "server/internal/modules/user/domain"
)

func TestJWTAccessToken(t *testing.T) {
	secret := "test-secret"
	user := authdomain.AuthenticatedUser{
		UserID:   uuid.New(),
		Username: "ana",
		Role:     userdomain.RoleAdmin,
	}

	t.Run("issues and verifies a token", func(t *testing.T) {
		issuer := NewJWTAccessTokenIssuer(secret, 15*time.Minute)
		verifier := NewJWTAccessTokenVerifier(secret)

		token, err := issuer.Issue(user)
		if err != nil {
			t.Fatalf("Issue() error = %v", err)
		}

		parsed, err := verifier.Verify(token)
		if err != nil {
			t.Fatalf("Verify() error = %v", err)
		}
		if parsed.UserID != user.UserID || parsed.Username != user.Username || parsed.Role != user.Role {
			t.Fatalf("unexpected parsed user: %#v", parsed)
		}
	})

	t.Run("rejects expired token", func(t *testing.T) {
		issuer := NewJWTAccessTokenIssuer(secret, -time.Minute)
		verifier := NewJWTAccessTokenVerifier(secret)

		token, err := issuer.Issue(user)
		if err != nil {
			t.Fatalf("Issue() error = %v", err)
		}

		if _, err := verifier.Verify(token); !errors.Is(err, ErrInvalidAccessToken) {
			t.Fatalf("error = %v, want ErrInvalidAccessToken", err)
		}
	})

	t.Run("rejects token with wrong secret", func(t *testing.T) {
		issuer := NewJWTAccessTokenIssuer(secret, 15*time.Minute)
		verifier := NewJWTAccessTokenVerifier("other-secret")

		token, err := issuer.Issue(user)
		if err != nil {
			t.Fatalf("Issue() error = %v", err)
		}

		if _, err := verifier.Verify(token); !errors.Is(err, ErrInvalidAccessToken) {
			t.Fatalf("error = %v, want ErrInvalidAccessToken", err)
		}
	})

	t.Run("rejects tampered token", func(t *testing.T) {
		issuer := NewJWTAccessTokenIssuer(secret, 15*time.Minute)
		verifier := NewJWTAccessTokenVerifier(secret)

		token, err := issuer.Issue(user)
		if err != nil {
			t.Fatalf("Issue() error = %v", err)
		}
		altered := token[:len(token)-1] + flipLastChar(token)

		if _, err := verifier.Verify(altered); !errors.Is(err, ErrInvalidAccessToken) {
			t.Fatalf("error = %v, want ErrInvalidAccessToken", err)
		}
	})
}

func flipLastChar(value string) string {
	if len(value) == 0 {
		return "x"
	}
	last := value[len(value)-1]
	if last == 'a' {
		return "b"
	}
	return "a"
}
