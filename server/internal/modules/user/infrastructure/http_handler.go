package infrastructure

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	userapplication "server/internal/modules/user/application"
	userdomain "server/internal/modules/user/domain"
	"server/internal/modules/user/ports"
	platformhttp "server/internal/platform/http"
)

type CreateUserUseCase interface {
	Execute(context.Context, userapplication.CreateUserCommand) (*userdomain.User, error)
}

type ListUsersUseCase interface {
	Execute(context.Context) ([]*userdomain.User, error)
}

type UpdateOwnUserUseCase interface {
	Execute(context.Context, uuid.UUID, userapplication.UpdateOwnUserCommand) (*userdomain.User, error)
}

type UpdateUserByAdminUseCase interface {
	Execute(context.Context, uuid.UUID, userapplication.UpdateUserByAdminCommand) (*userdomain.User, error)
}

type UserHandler struct {
	createUser      CreateUserUseCase
	listUsers       ListUsersUseCase
	updateOwnUser   UpdateOwnUserUseCase
	updateUserAdmin UpdateUserByAdminUseCase
	adminMiddleware func(http.Handler) http.Handler
}

// NewUserHandler wires the user use cases and admin middleware into a handler.
// It returns a handler ready to register its routes.
func NewUserHandler(
	createUser CreateUserUseCase,
	listUsers ListUsersUseCase,
	updateOwnUser UpdateOwnUserUseCase,
	updateUserAdmin UpdateUserByAdminUseCase,
	adminMiddleware func(http.Handler) http.Handler,
) *UserHandler {
	return &UserHandler{
		createUser:      createUser,
		listUsers:       listUsers,
		updateOwnUser:   updateOwnUser,
		updateUserAdmin: updateUserAdmin,
		adminMiddleware: adminMiddleware,
	}
}

// RegisterRoutes registers the user create, list and update endpoints on the mux.
// Admin-only routes are wrapped with the admin middleware.
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/users", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.HandleFunc("PATCH /api/v1/users/{id}", h.update)
	mux.Handle("GET /api/v1/admin/users", h.adminMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("PATCH /api/v1/admin/users/{id}", h.adminMiddleware(http.HandlerFunc(h.updateByAdmin)))
}

type createUserRequest struct {
	Name     string          `json:"name"`
	Username string          `json:"username"`
	Password string          `json:"password"`
	Role     userdomain.Role `json:"role"`
}

type updateUserRequest struct {
	Name     *string `json:"name"`
	Password *string `json:"password"`
}

type updateUserByAdminRequest struct {
	Name     *string          `json:"name"`
	Username *string          `json:"username"`
	Password *string          `json:"password"`
	Role     *userdomain.Role `json:"role"`
}

type userResponse struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	Username  string          `json:"username"`
	Role      userdomain.Role `json:"role"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
	CreatedBy string          `json:"created_by,omitempty"`
	UpdatedBy string          `json:"updated_by,omitempty"`
}

// create handles POST /api/v1/users and creates a user from the request body.
// It reads the actor from the context to set CreatedBy.
func (h *UserHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createUserRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	createdUser, err := h.createUser.Execute(r.Context(), userapplication.CreateUserCommand{
		Name:      request.Name,
		Username:  request.Username,
		Password:  request.Password,
		Role:      request.Role,
		CreatedBy: actor.UserID.String(),
	})
	if err != nil {
		writeUserError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusCreated, toUserResponse(createdUser))
}

// list handles GET /api/v1/admin/users and returns every registered user.
// It maps the domain users to the public response shape without passwords.
func (h *UserHandler) list(w http.ResponseWriter, r *http.Request) {
	users, err := h.listUsers.Execute(r.Context())
	if err != nil {
		writeUserError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toUserResponses(users))
}

// update handles PATCH /api/v1/users/{id} and updates the caller's own profile.
// It parses the id path value before executing the use case.
func (h *UserHandler) update(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del usuario no es válido"))
		return
	}

	var request updateUserRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	updatedUser, err := h.updateOwnUser.Execute(r.Context(), userID, userapplication.UpdateOwnUserCommand{
		Name:     request.Name,
		Password: request.Password,
	})
	if err != nil {
		writeUserError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toUserResponse(updatedUser))
}

// updateByAdmin handles PATCH /api/v1/admin/users/{id} and updates any user.
// It reads the actor from the context to set UpdatedBy.
func (h *UserHandler) updateByAdmin(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del usuario no es válido"))
		return
	}

	var request updateUserByAdminRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	updatedUser, err := h.updateUserAdmin.Execute(r.Context(), userID, userapplication.UpdateUserByAdminCommand{
		Name:      request.Name,
		Username:  request.Username,
		Password:  request.Password,
		Role:      request.Role,
		UpdatedBy: actor.UserID.String(),
	})
	if err != nil {
		writeUserError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toUserResponse(updatedUser))
}

// toUserResponse maps a domain user to the HTTP response shape.
// It formats timestamps in UTC and omits empty audit fields.
func toUserResponse(user *userdomain.User) userResponse {
	return userResponse{
		ID:        user.ID,
		Name:      user.Name,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt: user.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		CreatedBy: user.CreatedBy,
		UpdatedBy: user.UpdatedBy,
	}
}

// toUserResponses maps a list of domain users to HTTP response shapes.
// It returns an empty slice instead of null when there are no users.
func toUserResponses(users []*userdomain.User) []userResponse {
	responses := make([]userResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, toUserResponse(user))
	}
	return responses
}

// writeUserError maps domain and port errors to HTTP status codes.
// It falls back to 500 for unrecognized errors.
func writeUserError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrUsernameAlreadyUsed):
		status = http.StatusConflict
	case errors.Is(err, ports.ErrUserNotFound):
		status = http.StatusNotFound
	case errors.Is(err, userdomain.ErrInvalidID),
		errors.Is(err, userdomain.ErrInvalidName),
		errors.Is(err, userdomain.ErrInvalidUsername),
		errors.Is(err, userdomain.ErrInvalidPassword),
		errors.Is(err, userdomain.ErrInvalidRole),
		errors.Is(err, userdomain.ErrInvalidUpdate),
		errors.Is(err, userdomain.ErrInvalidUpdatedBy):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
