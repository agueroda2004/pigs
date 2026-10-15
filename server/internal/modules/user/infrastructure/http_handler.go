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

type DeleteUserUseCase interface {
	Execute(context.Context, uuid.UUID) error
}

type UserHandler struct {
	createUser      CreateUserUseCase
	listUsers       ListUsersUseCase
	updateOwnUser   UpdateOwnUserUseCase
	updateUserAdmin UpdateUserByAdminUseCase
	deleteUser      DeleteUserUseCase
	adminMiddleware func(http.Handler) http.Handler
}

// NewUserHandler wires the user use cases and admin middleware into a handler.
// It returns a handler ready to register its routes.
func NewUserHandler(
	createUser CreateUserUseCase,
	listUsers ListUsersUseCase,
	updateOwnUser UpdateOwnUserUseCase,
	updateUserAdmin UpdateUserByAdminUseCase,
	deleteUser DeleteUserUseCase,
	adminMiddleware func(http.Handler) http.Handler,
) *UserHandler {
	return &UserHandler{
		createUser:      createUser,
		listUsers:       listUsers,
		updateOwnUser:   updateOwnUser,
		updateUserAdmin: updateUserAdmin,
		deleteUser:      deleteUser,
		adminMiddleware: adminMiddleware,
	}
}

// RegisterRoutes registers the user create, list, update and delete endpoints on the mux.
// Admin-only routes are wrapped with the admin middleware.
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/users", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.HandleFunc("PATCH /api/v1/users/{id}", h.update)
	mux.Handle("GET /api/v1/admin/users", h.adminMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("PATCH /api/v1/admin/users/{id}", h.adminMiddleware(http.HandlerFunc(h.updateByAdmin)))
	mux.Handle("DELETE /api/v1/admin/users/{id}", h.adminMiddleware(http.HandlerFunc(h.delete)))
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
	Active   *bool            `json:"active"`
}

type userListResponse struct {
	ID       uuid.UUID       `json:"id"`
	Name     string          `json:"name"`
	Username string          `json:"username"`
	Role     userdomain.Role `json:"role"`
	Active   bool            `json:"active"`
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

	if _, err := h.createUser.Execute(r.Context(), userapplication.CreateUserCommand{
		Name:      request.Name,
		Username:  request.Username,
		Password:  request.Password,
		Role:      request.Role,
		CreatedBy: actor.UserID.String(),
	}); err != nil {
		writeUserError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// list handles GET /api/v1/admin/users and returns every registered user.
// It maps the domain users to a public response shape without passwords or audit fields.
func (h *UserHandler) list(w http.ResponseWriter, r *http.Request) {
	users, err := h.listUsers.Execute(r.Context())
	if err != nil {
		writeUserError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toUserListResponses(users))
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

	if _, err := h.updateOwnUser.Execute(r.Context(), userID, userapplication.UpdateOwnUserCommand{
		Name:     request.Name,
		Password: request.Password,
	}); err != nil {
		writeUserError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
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

	if _, err := h.updateUserAdmin.Execute(r.Context(), userID, userapplication.UpdateUserByAdminCommand{
		Name:      request.Name,
		Username:  request.Username,
		Password:  request.Password,
		Role:      request.Role,
		Active:    request.Active,
		UpdatedBy: actor.UserID.String(),
	}); err != nil {
		writeUserError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// delete handles DELETE /api/v1/admin/users/{id} and removes a user.
// It parses the id path value and maps a linked-records conflict to 409.
func (h *UserHandler) delete(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del usuario no es válido"))
		return
	}

	if err := h.deleteUser.Execute(r.Context(), userID); err != nil {
		writeUserError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// toUserListResponse maps a domain user to the list response shape.
// It exposes only the identity fields and omits every audit field.
func toUserListResponse(user *userdomain.User) userListResponse {
	return userListResponse{
		ID:       user.ID,
		Name:     user.Name,
		Username: user.Username,
		Role:     user.Role,
		Active:   user.Active,
	}
}

// toUserListResponses maps a list of domain users to list response shapes.
// It returns an empty slice instead of null when there are no users.
func toUserListResponses(users []*userdomain.User) []userListResponse {
	responses := make([]userListResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, toUserListResponse(user))
	}
	return responses
}

// writeUserError maps domain and port errors to HTTP status codes.
// It falls back to 500 for unrecognized errors.
func writeUserError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrUsernameAlreadyUsed),
		errors.Is(err, ports.ErrUserInUse):
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
