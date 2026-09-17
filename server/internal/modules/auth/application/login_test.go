package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	userdomain "server/internal/modules/user/domain"
	userports "server/internal/modules/user/ports"
)

func TestLoginServiceExecute(t *testing.T) {
	userID := uuid.New()
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	refreshTTL := 7 * 24 * time.Hour

	newService := func(users *fakeUserReader, verifier *fakePasswordVerifier, refresh *fakeRefreshTokenRepository) *LoginService {
		return NewLoginService(
			users,
			verifier,
			&fakeAccessTokenIssuer{token: "access-token"},
			&fakeTokenGenerator{token: "raw-refresh-token"},
			&fakeTokenHasher{hash: "refresh-hash"},
			refresh,
			func() time.Time { return now },
			refreshTTL,
		)
	}

	t.Run("logs in successfully", func(t *testing.T) {
		users := &fakeUserReader{findUser: testUser(userID)}
		verifier := &fakePasswordVerifier{}
		refresh := &fakeRefreshTokenRepository{}
		service := newService(users, verifier, refresh)

		result, err := service.Execute(context.Background(), LoginCommand{Username: "ana", Password: "secret"})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result.AccessToken != "access-token" || result.RefreshToken != "raw-refresh-token" {
			t.Fatalf("unexpected result: %#v", result)
		}
		if verifier.password != "secret" || verifier.hashed != "old-hash" {
			t.Fatalf("verifier received unexpected values: hashed=%q password=%q", verifier.hashed, verifier.password)
		}
		if len(refresh.created) != 1 || refresh.created[0].TokenHash != "refresh-hash" || refresh.created[0].UserID != userID || !refresh.created[0].ExpiresAt.Equal(now.Add(refreshTTL)) {
			t.Fatalf("unexpected created refresh token: %#v", refresh.created)
		}
		if !result.RefreshExpiresAt.Equal(now.Add(refreshTTL)) {
			t.Fatalf("unexpected expiry: %v", result.RefreshExpiresAt)
		}
	})

	t.Run("rejects unknown user", func(t *testing.T) {
		users := &fakeUserReader{findErr: userports.ErrUserNotFound}
		service := newService(users, &fakePasswordVerifier{}, &fakeRefreshTokenRepository{})

		_, err := service.Execute(context.Background(), LoginCommand{Username: "ana", Password: "secret"})

		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("error = %v, want ErrInvalidCredentials", err)
		}
	})

	t.Run("rejects wrong password", func(t *testing.T) {
		users := &fakeUserReader{findUser: testUser(userID)}
		verifier := &fakePasswordVerifier{err: errors.New("bcrypt mismatch")}
		service := newService(users, verifier, &fakeRefreshTokenRepository{})

		_, err := service.Execute(context.Background(), LoginCommand{Username: "ana", Password: "wrong"})

		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("error = %v, want ErrInvalidCredentials", err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		users := &fakeUserReader{findErr: expected}
		service := newService(users, &fakePasswordVerifier{}, &fakeRefreshTokenRepository{})

		_, err := service.Execute(context.Background(), LoginCommand{Username: "ana", Password: "secret"})

		if !errors.Is(err, expected) {
			t.Fatalf("error = %v, want %v", err, expected)
		}
	})

	t.Run("propagates issue error", func(t *testing.T) {
		expected := errors.New("issue failed")
		service := NewLoginService(
			&fakeUserReader{findUser: testUser(userID)},
			&fakePasswordVerifier{},
			&fakeAccessTokenIssuer{err: expected},
			&fakeTokenGenerator{},
			&fakeTokenHasher{},
			&fakeRefreshTokenRepository{},
			func() time.Time { return now },
			refreshTTL,
		)

		_, err := service.Execute(context.Background(), LoginCommand{Username: "ana", Password: "secret"})

		if !errors.Is(err, expected) {
			t.Fatalf("error = %v, want %v", err, expected)
		}
	})

	t.Run("propagates token generator error", func(t *testing.T) {
		expected := errors.New("generate failed")
		service := NewLoginService(
			&fakeUserReader{findUser: testUser(userID)},
			&fakePasswordVerifier{},
			&fakeAccessTokenIssuer{token: "access"},
			&fakeTokenGenerator{err: expected},
			&fakeTokenHasher{},
			&fakeRefreshTokenRepository{},
			func() time.Time { return now },
			refreshTTL,
		)

		_, err := service.Execute(context.Background(), LoginCommand{Username: "ana", Password: "secret"})

		if !errors.Is(err, expected) {
			t.Fatalf("error = %v, want %v", err, expected)
		}
	})
}

func testUser(id uuid.UUID) *userdomain.User {
	return &userdomain.User{
		ID: id, Name: "Ana", Username: "ana", Password: "old-hash",
		Role: userdomain.RoleUser, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
}
