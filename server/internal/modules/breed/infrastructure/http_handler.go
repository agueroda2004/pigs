package infrastructure

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	breedapplication "server/internal/modules/breed/application"
	breeddomain "server/internal/modules/breed/domain"
	"server/internal/modules/breed/ports"
	platformhttp "server/internal/platform/http"
)

type CreateBreedUseCase interface {
	Execute(context.Context, breedapplication.CreateBreedCommand) (*breeddomain.Breed, error)
}

type ListBreedsUseCase interface {
	Execute(context.Context) ([]*breeddomain.Breed, error)
}

type ListBreedOptionsUseCase interface {
	Execute(context.Context) ([]breeddomain.BreedOption, error)
}

type UpdateBreedUseCase interface {
	Execute(context.Context, uuid.UUID, breedapplication.UpdateBreedCommand) (*breeddomain.Breed, error)
}

type BreedHandler struct {
	createBreed      CreateBreedUseCase
	listBreeds       ListBreedsUseCase
	listBreedOptions ListBreedOptionsUseCase
	updateBreed      UpdateBreedUseCase
	authMiddleware   func(http.Handler) http.Handler
	adminMiddleware  func(http.Handler) http.Handler
}

// NewBreedHandler wires the breed use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewBreedHandler(
	createBreed CreateBreedUseCase,
	listBreeds ListBreedsUseCase,
	listOptions ListBreedOptionsUseCase,
	updateBreed UpdateBreedUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *BreedHandler {
	return &BreedHandler{
		createBreed:      createBreed,
		listBreeds:       listBreeds,
		listBreedOptions: listOptions,
		updateBreed:      updateBreed,
		authMiddleware:   authMiddleware,
		adminMiddleware:  adminMiddleware,
	}
}

// RegisterRoutes registers the breed create, list, options and update endpoints on the mux.
// Write routes are admin-only while the read routes only require authentication.
func (h *BreedHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/breeds", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/breeds", h.authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("GET /api/v1/breeds/options", h.authMiddleware(http.HandlerFunc(h.listOptions)))
	mux.Handle("PATCH /api/v1/breeds/{id}", h.adminMiddleware(http.HandlerFunc(h.update)))
}

type createBreedRequest struct {
	Name string `json:"name"`
}

type updateBreedRequest struct {
	Name   *string `json:"name"`
	Active *bool   `json:"active"`
}

type breedResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	CreatedBy uuid.UUID `json:"created_by"`
	UpdatedBy uuid.UUID `json:"updated_by"`
}

type breedOptionResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// create handles POST /api/v1/breeds and creates a breed from the request body.
// It reads the actor from the context to set CreatedBy.
func (h *BreedHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createBreedRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	createdBreed, err := h.createBreed.Execute(r.Context(), breedapplication.CreateBreedCommand{
		Name:      request.Name,
		CreatedBy: actor.UserID,
	})
	if err != nil {
		writeBreedError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusCreated, toBreedResponse(createdBreed))
}

// list handles GET /api/v1/breeds and returns every registered breed.
// It maps the domain breeds to the public response shape.
func (h *BreedHandler) list(w http.ResponseWriter, r *http.Request) {
	breeds, err := h.listBreeds.Execute(r.Context())
	if err != nil {
		writeBreedError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toBreedResponses(breeds))
}

// listOptions handles GET /api/v1/breeds/options and returns the active breeds.
// It maps the read model to a lightweight response with only id and name.
func (h *BreedHandler) listOptions(w http.ResponseWriter, r *http.Request) {
	options, err := h.listBreedOptions.Execute(r.Context())
	if err != nil {
		writeBreedError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toBreedOptionResponses(options))
}

// update handles PATCH /api/v1/breeds/{id} and applies the provided fields.
// It parses the id path value and reads the actor to set UpdatedBy.
func (h *BreedHandler) update(w http.ResponseWriter, r *http.Request) {
	breedID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador de la raza no es válido"))
		return
	}

	var request updateBreedRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	updatedBreed, err := h.updateBreed.Execute(r.Context(), breedID, breedapplication.UpdateBreedCommand{
		Name:      request.Name,
		Active:    request.Active,
		UpdatedBy: actor.UserID,
	})
	if err != nil {
		writeBreedError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toBreedResponse(updatedBreed))
}

// toBreedResponse maps a domain breed to the HTTP response shape.
// It formats timestamps in UTC.
func toBreedResponse(breed *breeddomain.Breed) breedResponse {
	return breedResponse{
		ID:        breed.ID,
		Name:      breed.Name,
		Active:    breed.Active,
		CreatedAt: breed.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt: breed.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		CreatedBy: breed.CreatedBy,
		UpdatedBy: breed.UpdatedBy,
	}
}

// toBreedResponses maps a list of domain breeds to HTTP response shapes.
// It returns an empty slice instead of null when there are no breeds.
func toBreedResponses(breeds []*breeddomain.Breed) []breedResponse {
	responses := make([]breedResponse, 0, len(breeds))
	for _, breed := range breeds {
		responses = append(responses, toBreedResponse(breed))
	}
	return responses
}

// toBreedOptionResponses maps breed options to their HTTP response shape.
// It returns an empty slice instead of null when there are no options.
func toBreedOptionResponses(options []breeddomain.BreedOption) []breedOptionResponse {
	responses := make([]breedOptionResponse, 0, len(options))
	for _, option := range options {
		responses = append(responses, breedOptionResponse{ID: option.ID, Name: option.Name})
	}
	return responses
}

// writeBreedError maps breed domain and port errors to HTTP status codes.
// It falls back to 500 for unrecognized errors.
func writeBreedError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrBreedNameAlreadyUsed):
		status = http.StatusConflict
	case errors.Is(err, ports.ErrBreedNotFound):
		status = http.StatusNotFound
	case errors.Is(err, breeddomain.ErrInvalidID),
		errors.Is(err, breeddomain.ErrInvalidName),
		errors.Is(err, breeddomain.ErrInvalidUpdate),
		errors.Is(err, breeddomain.ErrInvalidCreatedBy),
		errors.Is(err, breeddomain.ErrInvalidUpdatedBy):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
