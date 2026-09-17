package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewRefreshToken(t *testing.T) {
	userID := uuid.New()
	familyID := uuid.New()
	tokenID := uuid.New()
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	t.Run("creates a valid token", func(t *testing.T) {
		token, err := NewRefreshToken(tokenID, userID, familyID, "hash", now.Add(time.Hour), now)

		if err != nil {
			t.Fatalf("NewRefreshToken() error = %v", err)
		}
		if token.ID != tokenID || token.UserID != userID || token.FamilyID != familyID || token.TokenHash != "hash" {
			t.Fatalf("unexpected token: %#v", token)
		}
		if token.IsUsed || token.IsRevoked || !token.CreatedAt.Equal(now) || !token.ExpiresAt.Equal(now.Add(time.Hour)) {
			t.Fatalf("unexpected state: %#v", token)
		}
	})

	t.Run("rejects nil id", func(t *testing.T) {
		_, err := NewRefreshToken(uuid.Nil, userID, familyID, "hash", now.Add(time.Hour), now)
		if !errors.Is(err, ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})

	t.Run("rejects nil user id", func(t *testing.T) {
		_, err := NewRefreshToken(tokenID, uuid.Nil, familyID, "hash", now.Add(time.Hour), now)
		if !errors.Is(err, ErrInvalidUserID) {
			t.Fatalf("error = %v, want ErrInvalidUserID", err)
		}
	})

	t.Run("rejects nil family id", func(t *testing.T) {
		_, err := NewRefreshToken(tokenID, userID, uuid.Nil, "hash", now.Add(time.Hour), now)
		if !errors.Is(err, ErrInvalidFamilyID) {
			t.Fatalf("error = %v, want ErrInvalidFamilyID", err)
		}
	})

	t.Run("rejects empty hash", func(t *testing.T) {
		_, err := NewRefreshToken(tokenID, userID, familyID, "   ", now.Add(time.Hour), now)
		if !errors.Is(err, ErrInvalidTokenHash) {
			t.Fatalf("error = %v, want ErrInvalidTokenHash", err)
		}
	})

	t.Run("rejects past expiry", func(t *testing.T) {
		_, err := NewRefreshToken(tokenID, userID, familyID, "hash", now.Add(-time.Hour), now)
		if !errors.Is(err, ErrInvalidExpiry) {
			t.Fatalf("error = %v, want ErrInvalidExpiry", err)
		}
	})
}

func TestRefreshTokenState(t *testing.T) {
	userID := uuid.New()
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	newToken := func(expires time.Time) *RefreshToken {
		return &RefreshToken{
			ID: uuid.New(), UserID: userID, FamilyID: uuid.New(),
			TokenHash: "hash", ExpiresAt: expires, CreatedAt: now,
		}
	}

	t.Run("rotatable when fresh", func(t *testing.T) {
		token := newToken(now.Add(time.Hour))
		if err := token.CanBeRotated(now); err != nil {
			t.Fatalf("CanBeRotated() error = %v", err)
		}
		if token.IsExpiredAt(now) || token.WasReused() {
			t.Fatal("fresh token should not be expired or reused")
		}
	})

	t.Run("rejects expired token", func(t *testing.T) {
		token := newToken(now.Add(-time.Minute))
		if !errors.Is(token.CanBeRotated(now), ErrTokenExpired) {
			t.Fatal("expected ErrTokenExpired")
		}
		if !token.IsExpiredAt(now) {
			t.Fatal("expected IsExpiredAt to be true")
		}
	})

	t.Run("rejects revoked token", func(t *testing.T) {
		token := newToken(now.Add(time.Hour))
		token.Revoke()
		if !errors.Is(token.CanBeRotated(now), ErrTokenRevoked) || !token.WasReused() {
			t.Fatal("expected revoked token to be rejected and reused")
		}
	})

	t.Run("rejects used token and marks as used once", func(t *testing.T) {
		token := newToken(now.Add(time.Hour))
		if err := token.MarkAsUsed(); err != nil {
			t.Fatalf("MarkAsUsed() error = %v", err)
		}
		if !token.IsUsed {
			t.Fatal("expected token to be marked as used")
		}
		if !errors.Is(token.CanBeRotated(now), ErrTokenAlreadyUsed) || !token.WasReused() {
			t.Fatal("expected used token to be rejected and reused")
		}
		if !errors.Is(token.MarkAsUsed(), ErrTokenAlreadyUsed) {
			t.Fatal("expected second MarkAsUsed to fail")
		}
	})
}
