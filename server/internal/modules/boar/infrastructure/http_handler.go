package infrastructure

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	boarapplication "server/internal/modules/boar/application"
	boardomain "server/internal/modules/boar/domain"
	"server/internal/modules/boar/ports"
	platformhttp "server/internal/platform/http"
)

const dateLayout = "2006-01-02"

type CreateBoarUseCase interface {
	Execute(context.Context, boarapplication.CreateBoarCommand) (*boardomain.Boar, error)
}

type ListBoarsUseCase interface {
	Execute(context.Context, ports.BoarFilter) ([]*boardomain.Boar, error)
}

type UpdateBoarUseCase interface {
	Execute(context.Context, uuid.UUID, boarapplication.UpdateBoarCommand) (*boardomain.Boar, error)
}

type ListBoarDropdownUseCase interface {
	Execute(context.Context, bool, *boardomain.State) ([]boardomain.BoarDropdown, error)
}

type DeleteBoarUseCase interface {
	Execute(context.Context, uuid.UUID) error
}

type BoarHandler struct {
	createBoar       CreateBoarUseCase
	listBoars        ListBoarsUseCase
	listBoarDropdown ListBoarDropdownUseCase
	updateBoar       UpdateBoarUseCase
	deleteBoar       DeleteBoarUseCase
	authMiddleware   func(http.Handler) http.Handler
	adminMiddleware  func(http.Handler) http.Handler
}

// NewBoarHandler wires the boar use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewBoarHandler(
	createBoar CreateBoarUseCase,
	listBoars ListBoarsUseCase,
	listBoarDropdown ListBoarDropdownUseCase,
	updateBoar UpdateBoarUseCase,
	deleteBoar DeleteBoarUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *BoarHandler {
	return &BoarHandler{
		createBoar:       createBoar,
		listBoars:        listBoars,
		listBoarDropdown: listBoarDropdown,
		updateBoar:       updateBoar,
		deleteBoar:       deleteBoar,
		authMiddleware:   authMiddleware,
		adminMiddleware:  adminMiddleware,
	}
}

// RegisterRoutes registers the boar create, list, dropdown, update and delete endpoints.
// Write routes are admin-only while the read routes only require authentication.
func (h *BoarHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/boars", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/boars", h.authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("GET /api/v1/boars/dropdown", h.authMiddleware(http.HandlerFunc(h.listDropdown)))
	mux.Handle("PATCH /api/v1/boars/{id}", h.adminMiddleware(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /api/v1/boars/{id}", h.adminMiddleware(http.HandlerFunc(h.delete)))
}

type createBoarRequest struct {
	Code      string  `json:"code"`
	Location  *string `json:"location"`
	EntryDate string  `json:"entry_date"`
	BirthDate *string `json:"birth_date"`
	Note      *string `json:"note"`
	Origin    string  `json:"origin"`
	BreedID   string  `json:"breed_id"`
}

type updateBoarRequest struct {
	Code      *string `json:"code"`
	Location  *string `json:"location"`
	Active    *bool   `json:"active"`
	EntryDate *string `json:"entry_date"`
	BirthDate *string `json:"birth_date"`
	Note      *string `json:"note"`
	Origin    *string `json:"origin"`
	BreedID   *string `json:"breed_id"`
}

type boarSummaryResponse struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Location  *string   `json:"location"`
	Active    bool      `json:"active"`
	EntryDate string    `json:"entry_date"`
	BirthDate *string   `json:"birth_date"`
	Note      *string   `json:"note"`
	State     string    `json:"state"`
	Origin    string    `json:"origin"`
	BreedID   uuid.UUID `json:"breed_id"`
}

type boarDropdownResponse struct {
	ID     uuid.UUID `json:"id"`
	Code   string    `json:"code"`
	Active bool      `json:"active"`
}

// create handles POST /api/v1/boars and creates a boar from the request body.
// It reads the actor from the context to set CreatedBy and never accepts a state.
func (h *BoarHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createBoarRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	entryDate, err := time.Parse(dateLayout, request.EntryDate)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La fecha de ingreso no es válida"))
		return
	}

	birthDate, err := parseOptionalDate(request.BirthDate)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La fecha de nacimiento no es válida"))
		return
	}

	breedID, err := uuid.Parse(request.BreedID)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La raza no es válida"))
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if _, err := h.createBoar.Execute(r.Context(), boarapplication.CreateBoarCommand{
		Code:      request.Code,
		Location:  request.Location,
		EntryDate: entryDate,
		BirthDate: birthDate,
		Note:      request.Note,
		Origin:    boardomain.Origin(request.Origin),
		BreedID:   breedID,
		CreatedBy: actor.UserID,
	}); err != nil {
		writeBoarError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// list handles GET /api/v1/boars and returns the boars matching the query filter.
// It parses the optional code, breed_id, origin and active query parameters.
func (h *BoarHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseBoarFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	boars, err := h.listBoars.Execute(r.Context(), filter)
	if err != nil {
		writeBoarError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toBoarSummaries(boars))
}

// listDropdown handles GET /api/v1/boars/dropdown and returns boars for a selection list.
// It forwards the active flag and an optional state filter so callers may narrow the list.
func (h *BoarHandler) listDropdown(w http.ResponseWriter, r *http.Request) {
	active, err := parseActiveQuery(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	state, err := parseStateQuery(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	options, err := h.listBoarDropdown.Execute(r.Context(), active, state)
	if err != nil {
		writeBoarError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toBoarDropdownResponses(options))
}

// parseActiveQuery reads the optional active query parameter.
// It defaults to false (every boar) and rejects a malformed boolean.
func parseActiveQuery(r *http.Request) (bool, error) {
	active := strings.TrimSpace(r.URL.Query().Get("active"))
	if active == "" {
		return false, nil
	}
	parsedActive, err := strconv.ParseBool(active)
	if err != nil {
		return false, errors.New("El filtro de activo no es válido")
	}
	return parsedActive, nil
}

// parseStateQuery reads the optional state query parameter.
// It returns nil when empty (no state filter) and an error for an unknown state.
func parseStateQuery(r *http.Request) (*boardomain.State, error) {
	raw := strings.TrimSpace(r.URL.Query().Get("state"))
	if raw == "" {
		return nil, nil
	}
	state, err := boardomain.ParseState(raw)
	if err != nil {
		return nil, errors.New("El estado no es válido")
	}
	return &state, nil
}

// parseBoarFilter reads the optional list filters from the query string.
// It returns an error for a malformed breed_id, origin or active value.
func parseBoarFilter(r *http.Request) (ports.BoarFilter, error) {
	query := r.URL.Query()
	var filter ports.BoarFilter

	if code := strings.TrimSpace(query.Get("code")); code != "" {
		filter.Code = &code
	}

	if breed := strings.TrimSpace(query.Get("breed_id")); breed != "" {
		breedID, err := uuid.Parse(breed)
		if err != nil {
			return ports.BoarFilter{}, errors.New("El identificador de la raza no es válido")
		}
		filter.BreedID = &breedID
	}

	if origin := strings.TrimSpace(query.Get("origin")); origin != "" {
		parsedOrigin, err := boardomain.ParseOrigin(origin)
		if err != nil {
			return ports.BoarFilter{}, errors.New("El origen no es válido")
		}
		filter.Origin = &parsedOrigin
	}

	if active := strings.TrimSpace(query.Get("active")); active != "" {
		parsedActive, err := strconv.ParseBool(active)
		if err != nil {
			return ports.BoarFilter{}, errors.New("El filtro de activo no es válido")
		}
		filter.Active = &parsedActive
	}

	return filter, nil
}

// update handles PATCH /api/v1/boars/{id} and applies the provided fields.
// It parses the id path value, reads the actor and never accepts a state change.
func (h *BoarHandler) update(w http.ResponseWriter, r *http.Request) {
	boarID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del verraco no es válido"))
		return
	}

	var request updateBoarRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	entryDate, err := parseOptionalDate(request.EntryDate)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La fecha de ingreso no es válida"))
		return
	}

	var birthDate *time.Time
	clearBirthDate := false
	if request.BirthDate != nil {
		if *request.BirthDate == "" {
			clearBirthDate = true
		} else if birthDate, err = parseOptionalDate(request.BirthDate); err != nil {
			platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La fecha de nacimiento no es válida"))
			return
		}
	}

	var breedID *uuid.UUID
	if request.BreedID != nil {
		parsedBreedID, err := uuid.Parse(*request.BreedID)
		if err != nil {
			platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La raza no es válida"))
			return
		}
		breedID = &parsedBreedID
	}

	var origin *boardomain.Origin
	if request.Origin != nil {
		parsedOrigin := boardomain.Origin(*request.Origin)
		origin = &parsedOrigin
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if _, err := h.updateBoar.Execute(r.Context(), boarID, boarapplication.UpdateBoarCommand{
		Code:           request.Code,
		Location:       request.Location,
		Active:         request.Active,
		EntryDate:      entryDate,
		BirthDate:      birthDate,
		ClearBirthDate: clearBirthDate,
		Note:           request.Note,
		Origin:         origin,
		BreedID:        breedID,
		UpdatedBy:      actor.UserID,
	}); err != nil {
		writeBoarError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// delete handles DELETE /api/v1/boars/{id} and removes a boar.
// It parses the id path value and maps a linked-records conflict to 409.
func (h *BoarHandler) delete(w http.ResponseWriter, r *http.Request) {
	boarID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del verraco no es válido"))
		return
	}

	if err := h.deleteBoar.Execute(r.Context(), boarID); err != nil {
		writeBoarError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// parseOptionalDate parses an optional date pointer using the shared layout.
// It returns nil when value is nil and an error for an unparseable date.
func parseOptionalDate(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := time.Parse(dateLayout, *value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// toBoarSummary maps a domain boar to the HTTP response shape without audit fields.
// It formats the dates as YYYY-MM-DD.
func toBoarSummary(boar *boardomain.Boar) boarSummaryResponse {
	return boarSummaryResponse{
		ID:        boar.ID,
		Code:      boar.Code,
		Location:  boar.Location,
		Active:    boar.Active,
		EntryDate: boar.EntryDate.UTC().Format(dateLayout),
		BirthDate: formatOptionalDate(boar.BirthDate),
		Note:      boar.Note,
		State:     string(boar.State),
		Origin:    string(boar.Origin),
		BreedID:   boar.BreedID,
	}
}

// toBoarSummaries maps a list of domain boars to HTTP response shapes.
// It returns an empty slice instead of null when there are no boars.
func toBoarSummaries(boars []*boardomain.Boar) []boarSummaryResponse {
	responses := make([]boarSummaryResponse, 0, len(boars))
	for _, boar := range boars {
		responses = append(responses, toBoarSummary(boar))
	}
	return responses
}

// toBoarDropdownResponses maps boar dropdown items to their HTTP response shape.
// It returns an empty slice instead of null when there are no options.
func toBoarDropdownResponses(options []boardomain.BoarDropdown) []boarDropdownResponse {
	responses := make([]boarDropdownResponse, 0, len(options))
	for _, option := range options {
		responses = append(responses, boarDropdownResponse{ID: option.ID, Code: option.Code, Active: option.Active})
	}
	return responses
}

// formatOptionalDate formats an optional date as YYYY-MM-DD.
// It returns nil when the date pointer is nil.
func formatOptionalDate(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(dateLayout)
	return &formatted
}

// writeBoarError maps boar domain and port errors to HTTP status codes.
// It falls back to 500 for unrecognized errors.
func writeBoarError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrBoarCodeAlreadyUsed),
		errors.Is(err, ports.ErrBoarInUse):
		status = http.StatusConflict
	case errors.Is(err, ports.ErrBoarNotFound):
		status = http.StatusNotFound
	case errors.Is(err, boardomain.ErrInvalidID),
		errors.Is(err, boardomain.ErrInvalidCode),
		errors.Is(err, boardomain.ErrInvalidLocation),
		errors.Is(err, boardomain.ErrInvalidEntryDate),
		errors.Is(err, boardomain.ErrInvalidBirthDate),
		errors.Is(err, boardomain.ErrInvalidNote),
		errors.Is(err, boardomain.ErrInvalidState),
		errors.Is(err, boardomain.ErrInvalidOrigin),
		errors.Is(err, boardomain.ErrInvalidBreed),
		errors.Is(err, boardomain.ErrInvalidUpdate),
		errors.Is(err, boardomain.ErrInvalidCreatedBy),
		errors.Is(err, boardomain.ErrInvalidUpdatedBy):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
