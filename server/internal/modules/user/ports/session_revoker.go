package ports

import (
	"context"

	"github.com/google/uuid"
)

// SessionRevoker revokes all refresh-token sessions for a user.
type SessionRevoker interface {
	RevokeSessions(ctx context.Context, userID uuid.UUID) error
}