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
	sowapplication "server/internal/modules/sow/application"
	sowdomain "server/internal/modules/sow/domain"
	"server/internal/modules/sow/ports"
	platformhttp "server/internal/platform/http"
)

const dateLayout = "2006-01-02"

type CreateSowUseCase interface {
	Execute(context.Context, sowapplication.CreateSowCommand) (*sowdomain.Sow, error)
}

type ListSowsUseCase interface {
	Execute(context.Context, ports.SowFilter) ([]*sowdomain.Sow, error)
}

type ListSowOptionsUseCase interface {
	Execute(context.Context, *bool) ([]sowdomain.SowOption, error)
}

type UpdateSowUseCase interface {
	Execute(context.Context, uuid.UUID, sowapplication.UpdateSowCommand) (*sowdomain.Sow, error)
}

type SowHandler struct {
	createSow       CreateSowUseCase
	listSows        ListSowsUseCase
	listSowOptions  ListSowOptionsUseCase
	updateSow       UpdateSowUseCase
	authMiddleware  func(http.Handler) http.Handler
	adminMiddleware func(http.Handler) http.Handler
}

// NewSowHandler wires the sow use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewSowHandler(
	createSow CreateSowUseCase,
	listSows ListSowsUseCase,
	listSowOptions ListSowOptionsUseCase,
	updateSow UpdateSowUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *SowHandler {
	return &SowHandler{
		createSow:       createSow,
		listSows:        listSows,
		listSowOptions:  listSowOptions,
		updateSow:       updateSow,
		authMiddleware:  authMiddleware,
		adminMiddleware: adminMiddleware,
	}
}

// RegisterRoutes registers the sow create, list, options and update endpoints on the mux.
// Write routes are admin-only while the read routes only require authentication;
// no state route is exposed because the state is server-managed.
func (h *SowHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/sows", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/sows", h.authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("GET /api/v1/sows/options", h.authMiddleware(http.HandlerFunc(h.listOptions)))
	mux.Handle("PATCH /api/v1/sows/{id}", h.adminMiddleware(http.HandlerFunc(h.update)))
}

type createSowRequest struct {
	Code      string  `json:"code"`
	Location  *string `json:"location"`
	EntryDate string  `json:"entry_date"`
	BirthDate *string `json:"birth_date"`
	Note      *string `json:"note"`
	Origin    string  `json:"origin"`
	Parity    int     `json:"parity"`
	BreedID   string  `json:"breed_id"`
}

type updateSowRequest struct {
	Code      *string `json:"code"`
	Location  *string `json:"location"`
	Active    *bool   `json:"active"`
	EntryDate *string `json:"entry_date"`
	BirthDate *string `json:"birth_date"`
	Note      *string `json:"note"`
	Origin    *string `json:"origin"`
	BreedID   *string `json:"breed_id"`
}

type sowResponse struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Location  *string   `json:"location"`
	Active    bool      `json:"active"`
	EntryDate string    `json:"entry_date"`
	BirthDate *string   `json:"birth_date"`
	Note      *string   `json:"note"`
	State     string    `json:"state"`
	Origin    string    `json:"origin"`
	Parity    int       `json:"parity"`
	BreedID   uuid.UUID `json:"breed_id"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	CreatedBy uuid.UUID `json:"created_by"`
	UpdatedBy uuid.UUID `json:"updated_by"`
}

type sowOptionResponse struct {
	ID   uuid.UUID `json:"id"`
	Code string    `json:"code"`
}

// create handles POST /api/v1/sows and creates a sow from the request body.
// It reads the actor from the context to set CreatedBy and never accepts a state.
func (h *SowHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createSowRequest
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

	createdSow, err := h.createSow.Execute(r.Context(), sowapplication.CreateSowCommand{
		Code:      request.Code,
		Location:  request.Location,
		EntryDate: entryDate,
		BirthDate: birthDate,
		Note:      request.Note,
		Origin:    sowdomain.Origin(request.Origin),
		Parity:    request.Parity,
		BreedID:   breedID,
		CreatedBy: actor.UserID,
	})
	if err != nil {
		writeSowError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusCreated, toSowResponse(createdSow))
}

// list handles GET /api/v1/sows and returns the sows matching the query filter.
// It parses the optional code, breed_id, origin, active and state query parameters.
func (h *SowHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseSowFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	sows, err := h.listSows.Execute(r.Context(), filter)
	if err != nil {
		writeSowError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toSowResponses(sows))
}

// listOptions handles GET /api/v1/sows/options and returns the sow options.
// It forwards the optional active filter so callers may list all sows when empty.
func (h *SowHandler) listOptions(w http.ResponseWriter, r *http.Request) {
	active, err := parseSowOptionsActive(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	options, err := h.listSowOptions.Execute(r.Context(), active)
	if err != nil {
		writeSowError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toSowOptionResponses(options))
}

// parseSowOptionsActive reads the optional active query parameter.
// It returns nil when empty (no filter) and an error for a non-boolean value.
func parseSowOptionsActive(r *http.Request) (*bool, error) {
	active := strings.TrimSpace(r.URL.Query().Get("active"))
	if active == "" {
		return nil, nil
	}
	parsedActive, err := strconv.ParseBool(active)
	if err != nil {
		return nil, errors.New("El filtro de activo no es válido")
	}
	return &parsedActive, nil
}

// parseSowFilter reads the optional list filters from the query string.
// It returns an error for a malformed breed_id, origin, active or state value.
func parseSowFilter(r *http.Request) (ports.SowFilter, error) {
	query := r.URL.Query()
	var filter ports.SowFilter

	if code := strings.TrimSpace(query.Get("code")); code != "" {
		filter.Code = &code
	}

	if breed := strings.TrimSpace(query.Get("breed_id")); breed != "" {
		breedID, err := uuid.Parse(breed)
		if err != nil {
			return ports.SowFilter{}, errors.New("El identificador de la raza no es válido")
		}
		filter.BreedID = &breedID
	}

	if origin := strings.TrimSpace(query.Get("origin")); origin != "" {
		parsedOrigin, err := sowdomain.ParseOrigin(origin)
		if err != nil {
			return ports.SowFilter{}, errors.New("El origen no es válido")
		}
		filter.Origin = &parsedOrigin
	}

	if active := strings.TrimSpace(query.Get("active")); active != "" {
		parsedActive, err := strconv.ParseBool(active)
		if err != nil {
			return ports.SowFilter{}, errors.New("El filtro de activo no es válido")
		}
		filter.Active = &parsedActive
	}

	if state := strings.TrimSpace(query.Get("state")); state != "" {
		parsedState, err := sowdomain.ParseState(state)
		if err != nil {
			return ports.SowFilter{}, errors.New("El estado no es válido")
		}
		filter.State = &parsedState
	}

	return filter, nil
}

// update handles PATCH /api/v1/sows/{id} and applies the provided fields.
// It parses the id path value, reads the actor and never accepts state or parity.
func (h *SowHandler) update(w http.ResponseWriter, r *http.Request) {
	sowID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador de la cerda no es válido"))
		return
	}

	var request updateSowRequest
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

	var origin *sowdomain.Origin
	if request.Origin != nil {
		parsedOrigin := sowdomain.Origin(*request.Origin)
		origin = &parsedOrigin
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	updatedSow, err := h.updateSow.Execute(r.Context(), sowID, sowapplication.UpdateSowCommand{
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
	})
	if err != nil {
		writeSowError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toSowResponse(updatedSow))
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

// toSowResponse maps a domain sow to the HTTP response shape.
// It formats dates as YYYY-MM-DD and timestamps in UTC.
func toSowResponse(sow *sowdomain.Sow) sowResponse {
	return sowResponse{
		ID:        sow.ID,
		Code:      sow.Code,
		Location:  sow.Location,
		Active:    sow.Active,
		EntryDate: sow.EntryDate.UTC().Format(dateLayout),
		BirthDate: formatOptionalDate(sow.BirthDate),
		Note:      sow.Note,
		State:     string(sow.State),
		Origin:    string(sow.Origin),
		Parity:    sow.Parity,
		BreedID:   sow.BreedID,
		CreatedAt: sow.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt: sow.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		CreatedBy: sow.CreatedBy,
		UpdatedBy: sow.UpdatedBy,
	}
}

// toSowResponses maps a list of domain sows to HTTP response shapes.
// It returns an empty slice instead of null when there are no sows.
func toSowResponses(sows []*sowdomain.Sow) []sowResponse {
	responses := make([]sowResponse, 0, len(sows))
	for _, sow := range sows {
		responses = append(responses, toSowResponse(sow))
	}
	return responses
}

// toSowOptionResponses maps sow options to their HTTP response shape.
// It returns an empty slice instead of null when there are no options.
func toSowOptionResponses(options []sowdomain.SowOption) []sowOptionResponse {
	responses := make([]sowOptionResponse, 0, len(options))
	for _, option := range options {
		responses = append(responses, sowOptionResponse{ID: option.ID, Code: option.Code})
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

// writeSowError maps sow domain and port errors to HTTP status codes.
// It falls back to 500 for unrecognized errors.
func writeSowError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrSowCodeAlreadyUsed):
		status = http.StatusConflict
	case errors.Is(err, ports.ErrSowNotFound):
		status = http.StatusNotFound
	case errors.Is(err, sowdomain.ErrInvalidID),
		errors.Is(err, sowdomain.ErrInvalidCode),
		errors.Is(err, sowdomain.ErrInvalidLocation),
		errors.Is(err, sowdomain.ErrInvalidEntryDate),
		errors.Is(err, sowdomain.ErrInvalidBirthDate),
		errors.Is(err, sowdomain.ErrInvalidNote),
		errors.Is(err, sowdomain.ErrInvalidState),
		errors.Is(err, sowdomain.ErrInvalidOrigin),
		errors.Is(err, sowdomain.ErrInvalidBreed),
		errors.Is(err, sowdomain.ErrInvalidParity),
		errors.Is(err, sowdomain.ErrInvalidUpdate),
		errors.Is(err, sowdomain.ErrInvalidCreatedBy),
		errors.Is(err, sowdomain.ErrInvalidUpdatedBy):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
