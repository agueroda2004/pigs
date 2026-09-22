package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidID        = errors.New("El identificador del token es obligatorio")
	ErrInvalidUserID    = errors.New("El usuario del token es obligatorio")
	ErrInvalidFamilyID  = errors.New("La familia del token es obligatoria")
	ErrInvalidTokenHash = errors.New("El hash del token es obligatorio")
	ErrInvalidExpiry    = errors.New("La fecha de expiracion debe ser futura")
	ErrTokenExpired     = errors.New("El token de refresco ha expirado")
	ErrTokenRevoked     = errors.New("El token de refresco ha sido revocado")
	ErrTokenAlreadyUsed = errors.New("El token de refresco ya fue utilizado")
)

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	FamilyID  uuid.UUID
	TokenHash string
	IsUsed    bool
	IsRevoked bool
	ExpiresAt time.Time
	CreatedAt time.Time
}

// NewRefreshToken builds a refresh token after validating the ids, the token hash and the expiry.
// It requires a non-empty hash and an expiry strictly after now, and sets CreatedAt to now.
func NewRefreshToken(
	id uuid.UUID,
	userID uuid.UUID,
	familyID uuid.UUID,
	tokenHash string,
	expiresAt time.Time,
	now time.Time,
) (*RefreshToken, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidID
	}
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}
	if familyID == uuid.Nil {
		return nil, ErrInvalidFamilyID
	}
	if strings.TrimSpace(tokenHash) == "" {
		return nil, ErrInvalidTokenHash
	}
	if !expiresAt.After(now) {
		return nil, ErrInvalidExpiry
	}

	return &RefreshToken{
		ID:        id,
		UserID:    userID,
		FamilyID:  familyID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}, nil
}

// IsExpiredAt reports whether the token expiry is not after now.
// A nil receiver is treated as expired.
func (t *RefreshToken) IsExpiredAt(now time.Time) bool {
	if t == nil {
		return true
	}
	return !t.ExpiresAt.After(now)
}

// CanBeRotated checks whether the token may be rotated at now.
// It returns ErrTokenExpired, ErrTokenRevoked or ErrTokenAlreadyUsed when rotation is not allowed.
func (t *RefreshToken) CanBeRotated(now time.Time) error {
	if t == nil {
		return ErrTokenRevoked
	}
	if t.IsExpiredAt(now) {
		return ErrTokenExpired
	}
	if t.IsRevoked {
		return ErrTokenRevoked
	}
	if t.IsUsed {
		return ErrTokenAlreadyUsed
	}
	return nil
}

// MarkAsUsed flags the token as used so it cannot be rotated again.
// It returns ErrTokenRevoked or ErrTokenAlreadyUsed and leaves the token unchanged on error.
func (t *RefreshToken) MarkAsUsed() error {
	if t == nil {
		return ErrTokenRevoked
	}
	if t.IsRevoked {
		return ErrTokenRevoked
	}
	if t.IsUsed {
		return ErrTokenAlreadyUsed
	}
	t.IsUsed = true
	return nil
}

// Revoke marks the token as revoked.
// A nil receiver is ignored.
func (t *RefreshToken) Revoke() {
	if t == nil {
		return
	}
	t.IsRevoked = true
}

// WasReused reports whether the token was already used or revoked.
// A nil receiver is treated as reused.
func (t *RefreshToken) WasReused() bool {
	if t == nil {
		return true
	}
	return t.IsUsed || t.IsRevoked
}
