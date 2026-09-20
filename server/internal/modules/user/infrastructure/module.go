package infrastructure

import (
	"net/http"
	"time"

	userapplication "server/internal/modules/user/application"
	"server/internal/modules/user/ports"
)

type Module struct {
	CreateUser        *userapplication.CreateUserService
	UpdateOwnUser     *userapplication.UpdateOwnUserService
	UpdateUserByAdmin *userapplication.UpdateUserByAdminService
	Handler           *UserHandler
}

// NewModule assembles the user use cases and HTTP handler from its dependencies.
// It returns a module exposing the services and handler for wiring.
func NewModule(
	repository ports.UserRepository,
	hasher ports.PasswordHasher,
	clock func() time.Time,
	adminMiddleware func(http.Handler) http.Handler,
) *Module {
	createUser := userapplication.NewCreateUserService(repository, hasher, clock)
	updateOwnUser := userapplication.NewUpdateOwnUserService(repository, hasher, clock)
	updateUserByAdmin := userapplication.NewUpdateUserByAdminService(repository, hasher, clock)

	return &Module{
		CreateUser:        createUser,
		UpdateOwnUser:     updateOwnUser,
		UpdateUserByAdmin: updateUserByAdmin,
		Handler:           NewUserHandler(createUser, updateOwnUser, updateUserByAdmin, adminMiddleware),
	}
}
