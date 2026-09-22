package application

import (
	"context"
	"errors"

	"server/internal/modules/auth/ports"
)

type LogoutCommand struct {
	RefreshToken string
}

type LogoutService struct {
	refreshTokens ports.RefreshTokenRepository
	tokenHasher   ports.TokenHasher
}

// NewLogoutService builds the logout use case with its refresh token repository and hasher.
// It returns a service ready to execute LogoutCommand values.
func NewLogoutService(
	refreshTokens ports.RefreshTokenRepository,
	tokenHasher ports.TokenHasher,
) *LogoutService {
	return &LogoutService{
		refreshTokens: refreshTokens,
		tokenHasher:   tokenHasher,
	}
}

// Execute revokes the whole refresh token family for the given token.
// It succeeds silently when the token is not found.
func (s *LogoutService) Execute(ctx context.Context, command LogoutCommand) error {
	hash, err := s.tokenHasher.Hash(command.RefreshToken)
	if err != nil {
		return err
	}

	token, err := s.refreshTokens.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, ports.ErrRefreshTokenNotFound) {
			return nil
		}
		return err
	}

	return s.refreshTokens.RevokeFamily(ctx, token.FamilyID)
}
