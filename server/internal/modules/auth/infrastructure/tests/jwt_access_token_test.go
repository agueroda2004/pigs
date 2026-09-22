package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	authinfra "server/internal/modules/auth/infrastructure"
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
		issuer := authinfra.NewJWTAccessTokenIssuer(secret, 15*time.Minute)
		verifier := authinfra.NewJWTAccessTokenVerifier(secret)

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
		issuer := authinfra.NewJWTAccessTokenIssuer(secret, -time.Minute)
		verifier := authinfra.NewJWTAccessTokenVerifier(secret)

		token, err := issuer.Issue(user)
		if err != nil {
			t.Fatalf("Issue() error = %v", err)
		}

		if _, err := verifier.Verify(token); !errors.Is(err, authinfra.ErrInvalidAccessToken) {
			t.Fatalf("error = %v, want ErrInvalidAccessToken", err)
		}
	})

	t.Run("rejects token with wrong secret", func(t *testing.T) {
		issuer := authinfra.NewJWTAccessTokenIssuer(secret, 15*time.Minute)
		verifier := authinfra.NewJWTAccessTokenVerifier("other-secret")

		token, err := issuer.Issue(user)
		if err != nil {
			t.Fatalf("Issue() error = %v", err)
		}

		if _, err := verifier.Verify(token); !errors.Is(err, authinfra.ErrInvalidAccessToken) {
			t.Fatalf("error = %v, want ErrInvalidAccessToken", err)
		}
	})

	t.Run("rejects tampered token", func(t *testing.T) {
		issuer := authinfra.NewJWTAccessTokenIssuer(secret, 15*time.Minute)
		verifier := authinfra.NewJWTAccessTokenVerifier(secret)

		token, err := issuer.Issue(user)
		if err != nil {
			t.Fatalf("Issue() error = %v", err)
		}
		middle := len(token) / 2
		altered := token[:middle] + string(flipChar(token[middle])) + token[middle+1:]

		if _, err := verifier.Verify(altered); !errors.Is(err, authinfra.ErrInvalidAccessToken) {
			t.Fatalf("error = %v, want ErrInvalidAccessToken", err)
		}
	})
}

func flipChar(value byte) byte {
	if value == 'a' {
		return 'b'
	}
	return 'a'
}
