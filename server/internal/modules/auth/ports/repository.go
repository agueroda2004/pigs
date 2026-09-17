package ports

import (
	"context"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *authdomain.RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*authdomain.RefreshToken, error)
	MarkAsUsed(ctx context.Context, id uuid.UUID) error
	RevokeFamily(ctx context.Context, familyID uuid.UUID) error
	RevokeByHash(ctx context.Context, hash string) error
}
