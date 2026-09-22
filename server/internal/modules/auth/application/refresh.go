package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	"server/internal/modules/auth/ports"
)

type RefreshCommand struct {
	RefreshToken string
}

type RefreshResult struct {
	AccessToken      string
	RefreshToken     string
	RefreshExpiresAt time.Time
}

type RefreshService struct {
	refreshTokens  ports.RefreshTokenRepository
	users          ports.UserReader
	tokenHasher    ports.TokenHasher
	tokenGenerator ports.TokenGenerator
	accessIssuer   ports.AccessTokenIssuer
	clock          func() time.Time
	refreshTTL     time.Duration
}

// NewRefreshService builds the refresh use case with its repository, user reader and token helpers.
// It returns a service ready to execute RefreshCommand values.
func NewRefreshService(
	refreshTokens ports.RefreshTokenRepository,
	users ports.UserReader,
	tokenHasher ports.TokenHasher,
	tokenGenerator ports.TokenGenerator,
	accessIssuer ports.AccessTokenIssuer,
	clock func() time.Time,
	refreshTTL time.Duration,
) *RefreshService {
	return &RefreshService{
		refreshTokens:  refreshTokens,
		users:          users,
		tokenHasher:    tokenHasher,
		tokenGenerator: tokenGenerator,
		accessIssuer:   accessIssuer,
		clock:          clock,
		refreshTTL:     refreshTTL,
	}
}

// Execute rotates a refresh token and revokes the family when reuse is detected.
// It returns the matching domain error when the token is expired, revoked or already used.
func (s *RefreshService) Execute(ctx context.Context, command RefreshCommand) (*RefreshResult, error) {
	hash, err := s.tokenHasher.Hash(command.RefreshToken)
	if err != nil {
		return nil, err
	}

	token, err := s.refreshTokens.GetByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	now := s.clock()

	if err := token.CanBeRotated(now); err != nil {
		if token.WasReused() {
			if revokeErr := s.refreshTokens.RevokeFamily(ctx, token.FamilyID); revokeErr != nil {
				return nil, revokeErr
			}
		}
		return nil, err
	}

	if err := s.refreshTokens.MarkAsUsed(ctx, token.ID); err != nil {
		if errors.Is(err, ports.ErrRefreshTokenAlreadyUsed) {
			if revokeErr := s.refreshTokens.RevokeFamily(ctx, token.FamilyID); revokeErr != nil {
				return nil, revokeErr
			}
		}
		return nil, err
	}

	user, err := s.users.GetByID(ctx, token.UserID)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.accessIssuer.Issue(authdomain.AuthenticatedUser{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	})
	if err != nil {
		return nil, err
	}

	refreshTokenRaw, err := s.tokenGenerator.Generate()
	if err != nil {
		return nil, err
	}
	refreshHash, err := s.tokenHasher.Hash(refreshTokenRaw)
	if err != nil {
		return nil, err
	}

	newToken, err := authdomain.NewRefreshToken(
		uuid.New(),
		token.UserID,
		token.FamilyID,
		refreshHash,
		now.Add(s.refreshTTL),
		now,
	)
	if err != nil {
		return nil, err
	}

	if err := s.refreshTokens.Create(ctx, newToken); err != nil {
		return nil, err
	}

	return &RefreshResult{
		AccessToken:      accessToken,
		RefreshToken:     refreshTokenRaw,
		RefreshExpiresAt: newToken.ExpiresAt,
	}, nil
}
