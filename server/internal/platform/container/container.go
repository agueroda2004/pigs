package container

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	abortioninfrastructure "server/internal/modules/abortion/infrastructure"
	authinfrastructure "server/internal/modules/auth/infrastructure"
	boarinfrastructure "server/internal/modules/boar/infrastructure"
	boarremovalinfrastructure "server/internal/modules/boarremoval/infrastructure"
	breedinfrastructure "server/internal/modules/breed/infrastructure"
	farrowinginfrastructure "server/internal/modules/farrowing/infrastructure"
	medicationinfrastructure "server/internal/modules/medication/infrastructure"
	operatorinfrastructure "server/internal/modules/operator/infrastructure"
	pigletdeathinfrastructure "server/internal/modules/pigletdeath/infrastructure"
	pigletfosteringinfrastructure "server/internal/modules/pigletfostering/infrastructure"
	serviceinfrastructure "server/internal/modules/service/infrastructure"
	sowinfrastructure "server/internal/modules/sow/infrastructure"
	sowremovalinfrastructure "server/internal/modules/sowremoval/infrastructure"
	userinfrastructure "server/internal/modules/user/infrastructure"
	"server/internal/platform/config"
	"server/internal/platform/database"
)

type Container struct {
	DB              *pgxpool.Pool
	User            *userinfrastructure.Module
	Breed           *breedinfrastructure.Module
	Boar            *boarinfrastructure.Module
	BoarRemoval     *boarremovalinfrastructure.Module
	Sow             *sowinfrastructure.Module
	Operator        *operatorinfrastructure.Module
	Medication      *medicationinfrastructure.Module
	Service         *serviceinfrastructure.Module
	Abortion        *abortioninfrastructure.Module
	SowRemoval      *sowremovalinfrastructure.Module
	Farrowing       *farrowinginfrastructure.Module
	PigletDeath     *pigletdeathinfrastructure.Module
	PigletFostering *pigletfosteringinfrastructure.Module
	Auth            *authinfrastructure.Module
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

	// + === BOAR REMOVAL MODULE ===
	boarRemovalRepository := boarremovalinfrastructure.NewPostgresBoarRemovalRepository(db)

	// + === SOW MODULE ===
	sowRepository := sowinfrastructure.NewPostgresSowRepository(db)

	// + === OPERATOR MODULE ===
	operatorRepository := operatorinfrastructure.NewPostgresOperatorRepository(db)

	// + === MEDICATION MODULE ===
	medicationRepository := medicationinfrastructure.NewPostgresMedicationRepository(db)

	// + === SERVICE MODULE ===
	serviceRepository := serviceinfrastructure.NewPostgresServiceRepository(db)

	// + === ABORTION MODULE ===
	abortionRepository := abortioninfrastructure.NewPostgresAbortionRepository(db)

	// + === SOW REMOVAL MODULE ===
	sowRemovalRepository := sowremovalinfrastructure.NewPostgresSowRemovalRepository(db)

	// + === FARROWING MODULE ===
	farrowingRepository := farrowinginfrastructure.NewPostgresFarrowingRepository(db)

	// + === PIGLET DEATH MODULE ===
	pigletDeathRepository := pigletdeathinfrastructure.NewPostgresPigletDeathRepository(db)

	// + === PIGLET FOSTERING MODULE ===
	pigletFosteringRepository := pigletfosteringinfrastructure.NewPostgresPigletFosteringRepository(db)

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
		DB:              db,
		User:            userinfrastructure.NewModule(userRepository, hasher, clock, authModule.AdminMiddleware),
		Breed:           breedinfrastructure.NewModule(breedRepository, clock, authModule.AuthMiddleware, authModule.AdminMiddleware),
		Boar:            boarinfrastructure.NewModule(boarRepository, clock, authModule.AuthMiddleware, authModule.AdminMiddleware),
		BoarRemoval:     boarremovalinfrastructure.NewModule(boarRemovalRepository, clock, authModule.AuthMiddleware, authModule.AdminMiddleware),
		Sow:             sowinfrastructure.NewModule(sowRepository, clock, authModule.AuthMiddleware, authModule.AdminMiddleware),
		Operator:        operatorinfrastructure.NewModule(operatorRepository, clock, authModule.AuthMiddleware, authModule.AdminMiddleware),
		Medication:      medicationinfrastructure.NewModule(medicationRepository, clock, authModule.AuthMiddleware, authModule.AdminMiddleware),
		Service:         serviceinfrastructure.NewModule(serviceRepository, clock, authModule.AuthMiddleware, authModule.AdminMiddleware),
		Abortion:        abortioninfrastructure.NewModule(abortionRepository, clock, authModule.AuthMiddleware, authModule.AdminMiddleware),
		SowRemoval:      sowremovalinfrastructure.NewModule(sowRemovalRepository, clock, authModule.AuthMiddleware, authModule.AdminMiddleware),
		Farrowing:       farrowinginfrastructure.NewModule(farrowingRepository, clock, authModule.AuthMiddleware, authModule.AdminMiddleware),
		PigletDeath:     pigletdeathinfrastructure.NewModule(pigletDeathRepository, clock, authModule.AuthMiddleware, authModule.AdminMiddleware),
		PigletFostering: pigletfosteringinfrastructure.NewModule(pigletFosteringRepository, clock, authModule.AuthMiddleware, authModule.AdminMiddleware),
		Auth:            authModule,
	}, nil
}

func (c *Container) Close() {
	if c != nil && c.DB != nil {
		c.DB.Close()
	}
}
