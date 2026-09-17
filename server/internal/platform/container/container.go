package container

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	authinfrastructure "server/internal/modules/auth/infrastructure"
	userinfrastructure "server/internal/modules/user/infrastructure"
	"server/internal/platform/config"
	"server/internal/platform/database"
)

type Container struct {
	DB   *pgxpool.Pool
	User *userinfrastructure.Module
	Auth *authinfrastructure.Module
}

func New(ctx context.Context, applicationConfig config.Config) (*Container, error) {
	db, err := database.NewPostgresPool(ctx, applicationConfig.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// + === GLOBALS ===
	hasher := userinfrastructure.NewBcryptPasswordHasher(applicationConfig.BcryptCost)
	clock := time.Now

	// + === USER MODULE ===
	userRepository := userinfrastructure.NewPostgresUserRepository(db)

	// + === AUTH MODULE ===
	refreshTokenRepository := authinfrastructure.NewPostgresRefreshTokenRepository(db)
	tokenGenerator := authinfrastructure.NewRandomTokenGenerator()
	tokenHasher := authinfrastructure.NewSHA256TokenHasher()
	accessIssuer := authinfrastructure.NewJWTAccessTokenIssuer(
		applicationConfig.JWTSecret,
		applicationConfig.AccessTokenTTL,
	)
	accessVerifier := authinfrastructure.NewJWTAccessTokenVerifier(applicationConfig.JWTSecret)
	cookieConfig := authinfrastructure.CookieConfig{
		Secure:   applicationConfig.CookieSecure,
		SameSite: authinfrastructure.SameSiteFromString(applicationConfig.CookieSameSite),
	}

	authModule := authinfrastructure.NewModule(
		userRepository,
		refreshTokenRepository,
		hasher,
		accessIssuer,
		accessVerifier,
		tokenGenerator,
		tokenHasher,
		clock,
		applicationConfig.AccessTokenTTL,
		applicationConfig.RefreshTokenTTL,
		cookieConfig,
	)

	return &Container{
		DB:   db,
		User: userinfrastructure.NewModule(userRepository, hasher, clock, authModule.AdminMiddleware),
		Auth: authModule,
	}, nil
}

func (c *Container) Close() {
	if c != nil && c.DB != nil {
		c.DB.Close()
	}
}
