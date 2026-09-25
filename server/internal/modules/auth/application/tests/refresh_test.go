package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	authapplication "server/internal/modules/auth/application"
	authdomain "server/internal/modules/auth/domain"
	"server/internal/modules/auth/ports"
	userports "server/internal/modules/user/ports"
)

func TestRefreshServiceExecute(t *testing.T) {
	userID := uuid.New()
	familyID := uuid.New()
	tokenID := uuid.New()
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	refreshTTL := 7 * 24 * time.Hour

	newService := func(refresh *fakeRefreshTokenRepository, users *fakeUserReader, hasher *fakeTokenHasher) *authapplication.RefreshService {
		return authapplication.NewRefreshService(
			refresh,
			users,
			hasher,
			&fakeTokenGenerator{token: "new-refresh-token"},
			&fakeAccessTokenIssuer{token: "new-access-token"},
			func() time.Time { return now },
			refreshTTL,
		)
	}

	newToken := func(isUsed, isRevoked bool, expires time.Time) *authdomain.RefreshToken {
		return &authdomain.RefreshToken{
			ID: tokenID, UserID: userID, FamilyID: familyID, TokenHash: "hash",
			IsUsed: isUsed, IsRevoked: isRevoked, ExpiresAt: expires, CreatedAt: now,
		}
	}

	t.Run("rotates successfully keeping the family", func(t *testing.T) {
		refresh := &fakeRefreshTokenRepository{get: newToken(false, false, now.Add(time.Hour))}
		users := &fakeUserReader{getUser: testUser(userID)}
		hasher := &fakeTokenHasher{hash: "new-hash"}
		service := newService(refresh, users, hasher)

		result, err := service.Execute(context.Background(), authapplication.RefreshCommand{RefreshToken: "old-refresh-token"})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result.AccessToken != "new-access-token" || result.RefreshToken != "new-refresh-token" {
			t.Fatalf("unexpected result: %#v", result)
		}
		if refresh.usedID != tokenID {
			t.Fatalf("expected token %v to be marked as used, got %v", tokenID, refresh.usedID)
		}
		if len(refresh.created) != 1 || refresh.created[0].FamilyID != familyID || refresh.created[0].UserID != userID || refresh.created[0].TokenHash != "new-hash" {
			t.Fatalf("unexpected created token: %#v", refresh.created)
		}
		if len(hasher.tokens) != 2 || hasher.tokens[0] != "old-refresh-token" {
			t.Fatalf("unexpected hashed values: %#v", hasher.tokens)
		}
	})

	t.Run("rejects inactive user", func(t *testing.T) {
		inactive := testUser(userID)
		inactive.Active = false
		refresh := &fakeRefreshTokenRepository{get: newToken(false, false, now.Add(time.Hour))}
		service := newService(refresh, &fakeUserReader{getUser: inactive}, &fakeTokenHasher{})

		_, err := service.Execute(context.Background(), authapplication.RefreshCommand{RefreshToken: "token"})

		if !errors.Is(err, userports.ErrUserInactive) {
			t.Fatalf("error = %v, want ErrUserInactive", err)
		}
	})

	t.Run("propagates not found error", func(t *testing.T) {
		refresh := &fakeRefreshTokenRepository{getErr: ports.ErrRefreshTokenNotFound}
		service := newService(refresh, &fakeUserReader{}, &fakeTokenHasher{})

		_, err := service.Execute(context.Background(), authapplication.RefreshCommand{RefreshToken: "unknown"})

		if !errors.Is(err, ports.ErrRefreshTokenNotFound) {
			t.Fatalf("error = %v, want ErrRefreshTokenNotFound", err)
		}
	})

	t.Run("rejects expired token without revoking family", func(t *testing.T) {
		refresh := &fakeRefreshTokenRepository{get: newToken(false, false, now.Add(-time.Minute))}
		service := newService(refresh, &fakeUserReader{}, &fakeTokenHasher{})

		_, err := service.Execute(context.Background(), authapplication.RefreshCommand{RefreshToken: "expired"})

		if !errors.Is(err, authdomain.ErrTokenExpired) {
			t.Fatalf("error = %v, want ErrTokenExpired", err)
		}
		if refresh.revokedFamily != uuid.Nil {
			t.Fatalf("family should not be revoked, got %v", refresh.revokedFamily)
		}
	})

	t.Run("revokes family on reused token", func(t *testing.T) {
		for _, tt := range []struct {
			name   string
			isUsed bool
			revoke bool
			want   error
		}{
			{"used", true, false, authdomain.ErrTokenAlreadyUsed},
			{"revoked", false, true, authdomain.ErrTokenRevoked},
		} {
			t.Run(tt.name, func(t *testing.T) {
				refresh := &fakeRefreshTokenRepository{get: newToken(tt.isUsed, tt.revoke, now.Add(time.Hour))}
				service := newService(refresh, &fakeUserReader{}, &fakeTokenHasher{})

				_, err := service.Execute(context.Background(), authapplication.RefreshCommand{RefreshToken: "reused"})

				if !errors.Is(err, tt.want) {
					t.Fatalf("error = %v, want %v", err, tt.want)
				}
				if refresh.revokedFamily != familyID {
					t.Fatalf("expected family %v to be revoked, got %v", familyID, refresh.revokedFamily)
				}
			})
		}
	})

	t.Run("revokes family when token was already marked used", func(t *testing.T) {
		refresh := &fakeRefreshTokenRepository{
			get:    newToken(false, false, now.Add(time.Hour)),
			useErr: ports.ErrRefreshTokenAlreadyUsed,
		}
		service := newService(refresh, &fakeUserReader{}, &fakeTokenHasher{})

		_, err := service.Execute(context.Background(), authapplication.RefreshCommand{RefreshToken: "raced"})

		if !errors.Is(err, ports.ErrRefreshTokenAlreadyUsed) {
			t.Fatalf("error = %v, want ErrRefreshTokenAlreadyUsed", err)
		}
		if refresh.revokedFamily != familyID {
			t.Fatalf("expected family %v to be revoked, got %v", familyID, refresh.revokedFamily)
		}
	})

	t.Run("propagates user and issuer errors", func(t *testing.T) {
		userErr := errors.New("user failed")
		refresh := &fakeRefreshTokenRepository{get: newToken(false, false, now.Add(time.Hour))}
		service := newService(refresh, &fakeUserReader{getErr: userErr}, &fakeTokenHasher{})
		_, err := service.Execute(context.Background(), authapplication.RefreshCommand{RefreshToken: "token"})
		if !errors.Is(err, userErr) {
			t.Fatalf("error = %v, want %v", err, userErr)
		}

		issueErr := errors.New("issue failed")
		service = authapplication.NewRefreshService(
			refresh,
			&fakeUserReader{getUser: testUser(userID)},
			&fakeTokenHasher{},
			&fakeTokenGenerator{token: "new"},
			&fakeAccessTokenIssuer{err: issueErr},
			func() time.Time { return now },
			refreshTTL,
		)
		_, err = service.Execute(context.Background(), authapplication.RefreshCommand{RefreshToken: "token"})
		if !errors.Is(err, issueErr) {
			t.Fatalf("error = %v, want %v", err, issueErr)
		}
	})
}

func TestLogoutServiceExecute(t *testing.T) {
	userID := uuid.New()
	familyID := uuid.New()
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	t.Run("revokes the whole family", func(t *testing.T) {
		refresh := &fakeRefreshTokenRepository{
			get: &authdomain.RefreshToken{
				ID: uuid.New(), UserID: userID, FamilyID: familyID,
				TokenHash: "hash", ExpiresAt: now.Add(time.Hour), CreatedAt: now,
			},
		}
		hasher := &fakeTokenHasher{hash: "hash"}
		service := authapplication.NewLogoutService(refresh, hasher)

		err := service.Execute(context.Background(), authapplication.LogoutCommand{RefreshToken: "token"})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if refresh.revokedFamily != familyID {
			t.Fatalf("expected family %v to be revoked, got %v", familyID, refresh.revokedFamily)
		}
	})

	t.Run("succeeds when token is not found", func(t *testing.T) {
		refresh := &fakeRefreshTokenRepository{getErr: ports.ErrRefreshTokenNotFound}
		service := authapplication.NewLogoutService(refresh, &fakeTokenHasher{hash: "hash"})

		err := service.Execute(context.Background(), authapplication.LogoutCommand{RefreshToken: "unknown"})

		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		refresh := &fakeRefreshTokenRepository{getErr: expected}
		service := authapplication.NewLogoutService(refresh, &fakeTokenHasher{hash: "hash"})

		err := service.Execute(context.Background(), authapplication.LogoutCommand{RefreshToken: "token"})

		if !errors.Is(err, expected) {
			t.Fatalf("error = %v, want %v", err, expected)
		}
	})
}
