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

func WithAuthenticatedUser(ctx context.Context, user AuthenticatedUser) context.Context {
	return context.WithValue(ctx, contextKey{}, user)
}

func AuthenticatedUserFromContext(ctx context.Context) (AuthenticatedUser, bool) {
	user, ok := ctx.Value(contextKey{}).(AuthenticatedUser)
	return user, ok
}
