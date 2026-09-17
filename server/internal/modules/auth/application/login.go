package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	"server/internal/modules/auth/ports"
	userports "server/internal/modules/user/ports"
)

// + === ERRORS ===
var ErrInvalidCredentials = errors.New("Credenciales invalidas")

// + === TYPES ===
type LoginCommand struct {
	Username string
	Password string
}

type LoginResult struct {
	AccessToken      string
	RefreshToken     string
	RefreshExpiresAt time.Time
}

type LoginService struct {
	users            ports.UserReader
	passwordVerifier ports.PasswordVerifier
	accessIssuer     ports.AccessTokenIssuer
	tokenGenerator   ports.TokenGenerator
	tokenHasher      ports.TokenHasher
	refreshTokens    ports.RefreshTokenRepository
	clock            func() time.Time
	refreshTTL       time.Duration
}

// + === CONSTRUCTOR ===
func NewLoginService(
	users ports.UserReader,
	passwordVerifier ports.PasswordVerifier,
	accessIssuer ports.AccessTokenIssuer,
	tokenGenerator ports.TokenGenerator,
	tokenHasher ports.TokenHasher,
	refreshTokens ports.RefreshTokenRepository,
	clock func() time.Time,
	refreshTTL time.Duration,
) *LoginService {
	return &LoginService{
		users:            users,
		passwordVerifier: passwordVerifier,
		accessIssuer:     accessIssuer,
		tokenGenerator:   tokenGenerator,
		tokenHasher:      tokenHasher,
		refreshTokens:    refreshTokens,
		clock:            clock,
		refreshTTL:       refreshTTL,
	}
}

// + === METHODS ===
func (s *LoginService) Execute(ctx context.Context, command LoginCommand) (*LoginResult, error) {
	user, err := s.users.FindByUsername(ctx, command.Username)
	if err != nil {
		if errors.Is(err, userports.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := s.passwordVerifier.Verify(user.Password, command.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.accessIssuer.Issue(authdomain.AuthenticatedUser{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	})
	if err != nil {
		return nil, err
	}

	now := s.clock()

	refreshTokenRaw, err := s.tokenGenerator.Generate()
	if err != nil {
		return nil, err
	}
	refreshHash, err := s.tokenHasher.Hash(refreshTokenRaw)
	if err != nil {
		return nil, err
	}

	refreshToken, err := authdomain.NewRefreshToken(
		uuid.New(),
		user.ID,
		uuid.New(),
		refreshHash,
		now.Add(s.refreshTTL),
		now,
	)
	if err != nil {
		return nil, err
	}

	if err := s.refreshTokens.Create(ctx, refreshToken); err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:      accessToken,
		RefreshToken:     refreshTokenRaw,
		RefreshExpiresAt: refreshToken.ExpiresAt,
	}, nil
}
