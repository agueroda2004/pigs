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
	Execute(context.Context, ports.SowFilter, int) (sowapplication.SowPage, error)
}

type ListSowDropdownUseCase interface {
	Execute(context.Context, *bool, []sowdomain.State) ([]sowdomain.SowDropdown, error)
}

type UpdateSowUseCase interface {
	Execute(context.Context, uuid.UUID, sowapplication.UpdateSowCommand) (*sowdomain.Sow, error)
}

type SowHandler struct {
	createSow       CreateSowUseCase
	listSows        ListSowsUseCase
	listSowDropdown ListSowDropdownUseCase
	updateSow       UpdateSowUseCase
	authMiddleware  func(http.Handler) http.Handler
	adminMiddleware func(http.Handler) http.Handler
}

// NewSowHandler wires the sow use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewSowHandler(
	createSow CreateSowUseCase,
	listSows ListSowsUseCase,
	listSowDropdown ListSowDropdownUseCase,
	updateSow UpdateSowUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *SowHandler {
	return &SowHandler{
		createSow:       createSow,
		listSows:        listSows,
		listSowDropdown: listSowDropdown,
		updateSow:       updateSow,
		authMiddleware:  authMiddleware,
		adminMiddleware: adminMiddleware,
	}
}

// RegisterRoutes registers the sow create, list, dropdown and update endpoints on the mux.
// Write routes are admin-only while the read routes only require authentication;
// no state route is exposed because the state is server-managed.
func (h *SowHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/sows", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/sows", h.authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("GET /api/v1/sows/dropdown", h.authMiddleware(http.HandlerFunc(h.listDropdown)))
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
	EntryDate *string `json:"entry_date"`
	BirthDate *string `json:"birth_date"`
	Note      *string `json:"note"`
	Origin    *string `json:"origin"`
	BreedID   *string `json:"breed_id"`
	Parity    *int    `json:"parity"`
}

// sowSummaryResponse is the sow list shape without the audit fields.
type sowSummaryResponse struct {
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
}

// sowPageResponse is one page of sows together with its pagination metadata.
type sowPageResponse struct {
	Items      []sowSummaryResponse `json:"items"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

type sowDropdownResponse struct {
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

	if _, err := h.createSow.Execute(r.Context(), sowapplication.CreateSowCommand{
		Code:      request.Code,
		Location:  request.Location,
		EntryDate: entryDate,
		BirthDate: birthDate,
		Note:      request.Note,
		Origin:    sowdomain.Origin(request.Origin),
		Parity:    request.Parity,
		BreedID:   breedID,
		CreatedBy: actor.UserID,
	}); err != nil {
		writeSowError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// list handles GET /api/v1/sows and returns one page of sows matching the filter.
// It parses the optional code, breed_id, origin, active, state and page params.
func (h *SowHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseSowFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	page, err := parseSowPage(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	sows, err := h.listSows.Execute(r.Context(), filter, page)
	if err != nil {
		writeSowError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toSowPageResponse(sows))
}

// listDropdown handles GET /api/v1/sows/dropdown and returns the sow dropdown.
// It forwards the optional active filter and the list of states to filter by.
func (h *SowHandler) listDropdown(w http.ResponseWriter, r *http.Request) {
	active, err := parseSowDropdownActive(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	states, err := parseSowStates(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	options, err := h.listSowDropdown.Execute(r.Context(), active, states)
	if err != nil {
		writeSowError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toSowDropdownResponses(options))
}

// parseSowDropdownActive reads the optional active query parameter.
// It returns nil when empty (no filter) and an error for a non-boolean value.
func parseSowDropdownActive(r *http.Request) (*bool, error) {
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

// parseSowStates reads the optional repeated states query parameter and validates
// every value. It accepts comma-separated values inside each occurrence.
func parseSowStates(r *http.Request) ([]sowdomain.State, error) {
	rawStates := r.URL.Query()["states"]
	if len(rawStates) == 0 {
		return nil, nil
	}

	states := make([]sowdomain.State, 0, len(rawStates))
	for _, raw := range rawStates {
		for _, value := range strings.Split(raw, ",") {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			state, err := sowdomain.ParseState(value)
			if err != nil {
				return nil, errors.New("El estado no es válido")
			}
			states = append(states, state)
		}
	}
	if len(states) == 0 {
		return nil, nil
	}
	return states, nil
}

// parseSowPage reads the optional page query parameter.
// It defaults to page one and rejects non-positive or malformed values.
func parseSowPage(r *http.Request) (int, error) {
	raw := strings.TrimSpace(r.URL.Query().Get("page"))
	if raw == "" {
		return 1, nil
	}
	page, err := strconv.Atoi(raw)
	if err != nil || page < 1 {
		return 0, errors.New("La página no es válida")
	}
	return page, nil
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

	if _, err := h.updateSow.Execute(r.Context(), sowID, sowapplication.UpdateSowCommand{
		Code:           request.Code,
		Location:       request.Location,
		EntryDate:      entryDate,
		BirthDate:      birthDate,
		ClearBirthDate: clearBirthDate,
		Note:           request.Note,
		Origin:         origin,
		BreedID:        breedID,
		Parity:         request.Parity,
		UpdatedBy:      actor.UserID,
	}); err != nil {
		writeSowError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
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

// toSowSummary maps a domain sow to the list response shape without audit fields.
// It formats dates as YYYY-MM-DD.
func toSowSummary(sow *sowdomain.Sow) sowSummaryResponse {
	return sowSummaryResponse{
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
	}
}

// toSowPageResponse maps a page of domain sows to the paginated response shape.
// It returns an empty items slice instead of null when there are no sows.
func toSowPageResponse(page sowapplication.SowPage) sowPageResponse {
	items := make([]sowSummaryResponse, 0, len(page.Items))
	for _, sow := range page.Items {
		items = append(items, toSowSummary(sow))
	}
	return sowPageResponse{
		Items:      items,
		Total:      page.Total,
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalPages: page.TotalPages,
	}
}

// toSowDropdownResponses maps sow dropdown items to their HTTP response shape.
// It returns an empty slice instead of null when there are no items.
func toSowDropdownResponses(options []sowdomain.SowDropdown) []sowDropdownResponse {
	responses := make([]sowDropdownResponse, 0, len(options))
	for _, option := range options {
		responses = append(responses, sowDropdownResponse{ID: option.ID, Code: option.Code})
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
		errors.Is(err, sowdomain.ErrInvalidUpdatedBy),
		errors.Is(err, sowdomain.ErrInactiveSowCannotUpdateDates),
		errors.Is(err, sowdomain.ErrParityRequiresAliveState),
		errors.Is(err, sowdomain.ErrEntryDateAfterService):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
