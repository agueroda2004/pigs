package container

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	userapplication "server/internal/modules/user/application"
	userinfrastructure "server/internal/modules/user/infrastructure"
	"server/internal/platform/config"
	"server/internal/platform/database"
)

type Container struct {
	DB *pgxpool.Pool

	CreateUser    *userapplication.CreateUserService
	UpdateOwnUser *userapplication.UpdateOwnUserService
}

func New(ctx context.Context, applicationConfig config.Config) (*Container, error) {
	db, err := database.NewPostgresPool(ctx, applicationConfig.DatabaseURL)
	if err != nil {
		return nil, err
	}

	repository := userinfrastructure.NewPostgresUserRepository(db)
	hasher := userinfrastructure.NewBcryptPasswordHasher(applicationConfig.BcryptCost)
	clock := time.Now

	return &Container{
		DB:            db,
		CreateUser:    userapplication.NewCreateUserService(repository, hasher, clock),
		UpdateOwnUser: userapplication.NewUpdateOwnUserService(repository, hasher, clock),
	}, nil
}

func (c *Container) Close() {
	if c != nil && c.DB != nil {
		c.DB.Close()
	}
}
