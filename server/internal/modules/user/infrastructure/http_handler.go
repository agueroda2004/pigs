package infrastructure

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	userapplication "server/internal/modules/user/application"
	userdomain "server/internal/modules/user/domain"
	"server/internal/modules/user/ports"
)

const maxRequestBodySize = 1 << 20

type UserHandler struct {
	createUser    *userapplication.CreateUserService
	updateOwnUser *userapplication.UpdateOwnUserService
}

func NewUserHandler(
	createUser *userapplication.CreateUserService,
	updateOwnUser *userapplication.UpdateOwnUserService,
) *UserHandler {
	return &UserHandler{
		createUser:    createUser,
		updateOwnUser: updateOwnUser,
	}
}

func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/users", h.create)
	mux.HandleFunc("PATCH /api/v1/users/{id}", h.update)
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

type errorResponse struct {
	Error string `json:"error"`
}

func (h *UserHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createUserRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	createdBy := strings.TrimSpace(r.Header.Get("X-Actor-ID"))
	if createdBy == "" {
		createdBy = "local"
	}

	createdUser, err := h.createUser.Execute(r.Context(), userapplication.CreateUserCommand{
		Name:      request.Name,
		Username:  request.Username,
		Password:  request.Password,
		Role:      request.Role,
		CreatedBy: createdBy,
	})
	if err != nil {
		writeUserError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toUserResponse(createdUser))
}

func (h *UserHandler) update(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("El identificador del usuario no es válido"))
		return
	}

	var request updateUserRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
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

	writeJSON(w, http.StatusOK, toUserResponse(updatedUser))
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return errors.New("El cuerpo de la solicitud no es válido")
	}
	return nil
}

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
	writeError(w, status, err)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, errorResponse{Error: err.Error()})
}
