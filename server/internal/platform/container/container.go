package container

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	authinfrastructure "server/internal/modules/auth/infrastructure"
	boarinfrastructure "server/internal/modules/boar/infrastructure"
	breedinfrastructure "server/internal/modules/breed/infrastructure"
	operatorinfrastructure "server/internal/modules/operator/infrastructure"
	serviceinfrastructure "server/internal/modules/service/infrastructure"
	sowinfrastructure "server/internal/modules/sow/infrastructure"
	userinfrastructure "server/internal/modules/user/infrastructure"
	"server/internal/platform/config"
	"server/internal/platform/database"
)

type Container struct {
	DB       *pgxpool.Pool
	User     *userinfrastructure.Module
	Breed    *breedinfrastructure.Module
	Boar     *boarinfrastructure.Module
	Sow      *sowinfrastructure.Module
	Operator *operatorinfrastructure.Module
	Service  *serviceinfrastructure.Module
	Auth     *authinfrastructure.Module
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

	// + === BREED MODULE ===
	breedRepository := breedinfrastructure.NewPostgresBreedRepository(db)

	// + === BOAR MODULE ===
	boarRepository := boarinfrastructure.NewPostgresBoarRepository(db)

	// + === SOW MODULE ===
	sowRepository := sowinfrastructure.NewPostgresSowRepository(db)

	// + === OPERATOR MODULE ===
	operatorRepository := operatorinfrastructure.NewPostgresOperatorRepository(db)

	// + === SERVICE MODULE ===
	serviceRepository := serviceinfrastructure.NewPostgresServiceRepository(db)

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
		DB:       db,
		User:     userinfrastructure.NewModule(userRepository, hasher, clock, authModule.AdminMiddleware),
		Breed:    breedinfrastructure.NewModule(breedRepository, clock, authModule.AuthMiddleware, authModule.AdminMiddleware),
		Boar:     boarinfrastructure.NewModule(boarRepository, clock, authModule.AuthMiddleware, authModule.AdminMiddleware),
		Sow:      sowinfrastructure.NewModule(sowRepository, clock, authModule.AuthMiddleware, authModule.AdminMiddleware),
		Operator: operatorinfrastructure.NewModule(operatorRepository, clock, authModule.AuthMiddleware, authModule.AdminMiddleware),
		Service:  serviceinfrastructure.NewModule(serviceRepository, clock, authModule.AuthMiddleware, authModule.AdminMiddleware),
		Auth:     authModule,
	}, nil
}

func (c *Container) Close() {
	if c != nil && c.DB != nil {
		c.DB.Close()
	}
}
