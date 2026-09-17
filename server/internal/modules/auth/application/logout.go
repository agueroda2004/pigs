package application

import (
	"context"
	"errors"

	"server/internal/modules/auth/ports"
)

// + === TYPES ===
type LogoutCommand struct {
	RefreshToken string
}

type LogoutService struct {
	refreshTokens ports.RefreshTokenRepository
	tokenHasher   ports.TokenHasher
}

// + === CONSTRUCTOR ===
func NewLogoutService(
	refreshTokens ports.RefreshTokenRepository,
	tokenHasher ports.TokenHasher,
) *LogoutService {
	return &LogoutService{
		refreshTokens: refreshTokens,
		tokenHasher:   tokenHasher,
	}
}

// + === METHODS ===
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
