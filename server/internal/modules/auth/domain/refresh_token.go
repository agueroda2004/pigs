package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// + === ERRORS ===
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

// + === TYPE ===
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

// + === CONSTRUCTOR ===
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

// + === METHODS ===
func (t *RefreshToken) IsExpiredAt(now time.Time) bool {
	if t == nil {
		return true
	}
	return !t.ExpiresAt.After(now)
}

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

func (t *RefreshToken) Revoke() {
	if t == nil {
		return
	}
	t.IsRevoked = true
}

func (t *RefreshToken) WasReused() bool {
	if t == nil {
		return true
	}
	return t.IsUsed || t.IsRevoked
}
