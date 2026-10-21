package tests

import (
	"errors"
	"strings"
	"testing"

	authdomain "server/internal/modules/auth/domain"
)

func TestNewLoginCredentials(t *testing.T) {
	t.Run("trims the username and keeps the password", func(t *testing.T) {
		credentials, err := authdomain.NewLoginCredentials("  ana  ", "secret")

		if err != nil {
			t.Fatalf("NewLoginCredentials() error = %v", err)
		}
		if credentials.Username != "ana" || credentials.Password != "secret" {
			t.Fatalf("unexpected credentials: %#v", credentials)
		}
	})

	t.Run("accepts both fields at exactly the max length", func(t *testing.T) {
		credentials, err := authdomain.NewLoginCredentials(strings.Repeat("a", 50), strings.Repeat("b", 50))

		if err != nil {
			t.Fatalf("NewLoginCredentials() error = %v", err)
		}
		if len(credentials.Username) != 50 || len(credentials.Password) != 50 {
			t.Fatalf("unexpected credentials: %#v", credentials)
		}
	})

	t.Run("rejects empty values", func(t *testing.T) {
		if _, err := authdomain.NewLoginCredentials("   ", "secret"); !errors.Is(err, authdomain.ErrInvalidUsername) {
			t.Fatalf("error = %v, want ErrInvalidUsername", err)
		}

		if _, err := authdomain.NewLoginCredentials("ana", ""); !errors.Is(err, authdomain.ErrInvalidPassword) {
			t.Fatalf("error = %v, want ErrInvalidPassword", err)
		}
	})

	t.Run("rejects values over the max length", func(t *testing.T) {
		if _, err := authdomain.NewLoginCredentials(strings.Repeat("a", 51), "secret"); !errors.Is(err, authdomain.ErrInvalidUsername) {
			t.Fatalf("error = %v, want ErrInvalidUsername", err)
		}

		if _, err := authdomain.NewLoginCredentials("ana", strings.Repeat("b", 51)); !errors.Is(err, authdomain.ErrInvalidPassword) {
			t.Fatalf("error = %v, want ErrInvalidPassword", err)
		}
	})
}
