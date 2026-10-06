package infrastructure

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

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
	Execute(context.Context, ports.BreedFilter) ([]*breeddomain.Breed, error)
}

type ListBreedDropdownUseCase interface {
	Execute(context.Context, bool) ([]breeddomain.BreedDropdown, error)
}

type UpdateBreedUseCase interface {
	Execute(context.Context, uuid.UUID, breedapplication.UpdateBreedCommand) (*breeddomain.Breed, error)
}

type DeleteBreedUseCase interface {
	Execute(context.Context, uuid.UUID) error
}

type BreedHandler struct {
	createBreed       CreateBreedUseCase
	listBreeds        ListBreedsUseCase
	listBreedDropdown ListBreedDropdownUseCase
	updateBreed       UpdateBreedUseCase
	deleteBreed       DeleteBreedUseCase
	authMiddleware    func(http.Handler) http.Handler
	adminMiddleware   func(http.Handler) http.Handler
}

// NewBreedHandler wires the breed use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewBreedHandler(
	createBreed CreateBreedUseCase,
	listBreeds ListBreedsUseCase,
	listDropdown ListBreedDropdownUseCase,
	updateBreed UpdateBreedUseCase,
	deleteBreed DeleteBreedUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *BreedHandler {
	return &BreedHandler{
		createBreed:       createBreed,
		listBreeds:        listBreeds,
		listBreedDropdown: listDropdown,
		updateBreed:       updateBreed,
		deleteBreed:       deleteBreed,
		authMiddleware:    authMiddleware,
		adminMiddleware:   adminMiddleware,
	}
}

// RegisterRoutes registers the breed create, list, dropdown, update and delete endpoints on the mux.
// Write routes are admin-only while the read routes only require authentication.
func (h *BreedHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/breeds", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/breeds", h.authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("GET /api/v1/breeds/dropdown", h.authMiddleware(http.HandlerFunc(h.listDropdown)))
	mux.Handle("PATCH /api/v1/breeds/{id}", h.adminMiddleware(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /api/v1/breeds/{id}", h.adminMiddleware(http.HandlerFunc(h.delete)))
}

type createBreedRequest struct {
	Name string `json:"name"`
}

type updateBreedRequest struct {
	Name   *string `json:"name"`
	Active *bool   `json:"active"`
}

type breedSummaryResponse struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Active bool      `json:"active"`
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

	if _, err := h.createBreed.Execute(r.Context(), breedapplication.CreateBreedCommand{
		Name:      request.Name,
		CreatedBy: actor.UserID,
	}); err != nil {
		writeBreedError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// list handles GET /api/v1/breeds and returns the breeds matching the filter.
// It parses the optional name and active query parameters and omits the audit fields.
func (h *BreedHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseBreedFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	breeds, err := h.listBreeds.Execute(r.Context(), filter)
	if err != nil {
		writeBreedError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toBreedSummaries(breeds))
}

// listDropdown handles GET /api/v1/breeds/dropdown and returns breeds for a selection list.
// The active query parameter (default false) returns only active breeds or every breed.
func (h *BreedHandler) listDropdown(w http.ResponseWriter, r *http.Request) {
	active, err := parseActiveQuery(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	options, err := h.listBreedDropdown.Execute(r.Context(), active)
	if err != nil {
		writeBreedError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toBreedDropdownResponses(options))
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

	if _, err := h.updateBreed.Execute(r.Context(), breedID, breedapplication.UpdateBreedCommand{
		Name:      request.Name,
		Active:    request.Active,
		UpdatedBy: actor.UserID,
	}); err != nil {
		writeBreedError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// delete handles DELETE /api/v1/breeds/{id} and removes a breed.
// It parses the id path value and maps a linked-records conflict to 409.
func (h *BreedHandler) delete(w http.ResponseWriter, r *http.Request) {
	breedID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador de la raza no es válido"))
		return
	}

	if err := h.deleteBreed.Execute(r.Context(), breedID); err != nil {
		writeBreedError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// parseBreedFilter reads the optional name and active filters from the query string.
// It returns an error when the active value is not a valid boolean.
func parseBreedFilter(r *http.Request) (ports.BreedFilter, error) {
	query := r.URL.Query()
	var filter ports.BreedFilter

	if name := strings.TrimSpace(query.Get("name")); name != "" {
		filter.Name = &name
	}

	if raw := strings.TrimSpace(query.Get("active")); raw != "" {
		active, err := strconv.ParseBool(raw)
		if err != nil {
			return ports.BreedFilter{}, errors.New("El filtro activo no es válido")
		}
		filter.Active = &active
	}

	return filter, nil
}

// parseActiveQuery reads the optional active query parameter.
// It defaults to false (every breed) and rejects a malformed boolean.
func parseActiveQuery(r *http.Request) (bool, error) {
	raw := strings.TrimSpace(r.URL.Query().Get("active"))
	if raw == "" {
		return false, nil
	}
	active, err := strconv.ParseBool(raw)
	if err != nil {
		return false, errors.New("El filtro activo no es válido")
	}
	return active, nil
}

// toBreedSummaries maps domain breeds to the lightweight list response shape.
// It only exposes the identifier, name and active flag.
func toBreedSummaries(breeds []*breeddomain.Breed) []breedSummaryResponse {
	responses := make([]breedSummaryResponse, 0, len(breeds))
	for _, breed := range breeds {
		responses = append(responses, breedSummaryResponse{ID: breed.ID, Name: breed.Name, Active: breed.Active})
	}
	return responses
}

// toBreedDropdownResponses maps breed dropdown items to their HTTP response shape.
// It returns an empty slice instead of null when there are no items.
func toBreedDropdownResponses(options []breeddomain.BreedDropdown) []breedSummaryResponse {
	responses := make([]breedSummaryResponse, 0, len(options))
	for _, option := range options {
		responses = append(responses, breedSummaryResponse{ID: option.ID, Name: option.Name, Active: option.Active})
	}
	return responses
}

// writeBreedError maps breed domain and port errors to HTTP status codes.
// It falls back to 500 for unrecognized errors.
func writeBreedError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrBreedNameAlreadyUsed),
		errors.Is(err, ports.ErrBreedInUse):
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
