package auth

import (
	"context"

	"github.com/google/uuid"

	userdomain "server/internal/modules/user/domain"
)

type AuthenticatedUser struct {
	UserID   uuid.UUID
	Username string
	Role     userdomain.Role
}

type contextKey struct{}

// WithAuthenticatedUser stores the authenticated user in the context.
// It returns a new context carrying the user for downstream handlers.
func WithAuthenticatedUser(ctx context.Context, user AuthenticatedUser) context.Context {
	return context.WithValue(ctx, contextKey{}, user)
}

// AuthenticatedUserFromContext extracts the authenticated user from the context.
// It reports false when the context does not carry an authenticated user.
func AuthenticatedUserFromContext(ctx context.Context) (AuthenticatedUser, bool) {
	user, ok := ctx.Value(contextKey{}).(AuthenticatedUser)
	return user, ok
}
