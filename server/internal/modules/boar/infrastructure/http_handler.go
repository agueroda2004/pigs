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

type BoarHandler struct {
	createBoar      CreateBoarUseCase
	listBoars       ListBoarsUseCase
	updateBoar      UpdateBoarUseCase
	authMiddleware  func(http.Handler) http.Handler
	adminMiddleware func(http.Handler) http.Handler
}

// NewBoarHandler wires the boar use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewBoarHandler(
	createBoar CreateBoarUseCase,
	listBoars ListBoarsUseCase,
	updateBoar UpdateBoarUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *BoarHandler {
	return &BoarHandler{
		createBoar:      createBoar,
		listBoars:       listBoars,
		updateBoar:      updateBoar,
		authMiddleware:  authMiddleware,
		adminMiddleware: adminMiddleware,
	}
}

// RegisterRoutes registers the boar create, list and update endpoints on the mux.
// Write routes are admin-only while the list route only requires authentication.
func (h *BoarHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/boars", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/boars", h.authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("PATCH /api/v1/boars/{id}", h.adminMiddleware(http.HandlerFunc(h.update)))
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

type boarResponse struct {
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
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	CreatedBy uuid.UUID `json:"created_by"`
	UpdatedBy uuid.UUID `json:"updated_by"`
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

	createdBoar, err := h.createBoar.Execute(r.Context(), boarapplication.CreateBoarCommand{
		Code:      request.Code,
		Location:  request.Location,
		EntryDate: entryDate,
		BirthDate: birthDate,
		Note:      request.Note,
		Origin:    boardomain.Origin(request.Origin),
		BreedID:   breedID,
		CreatedBy: actor.UserID,
	})
	if err != nil {
		writeBoarError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusCreated, toBoarResponse(createdBoar))
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

	platformhttp.WriteJSON(w, http.StatusOK, toBoarResponses(boars))
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

	updatedBoar, err := h.updateBoar.Execute(r.Context(), boarID, boarapplication.UpdateBoarCommand{
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
		writeBoarError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toBoarResponse(updatedBoar))
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

// toBoarResponse maps a domain boar to the HTTP response shape.
// It formats dates as YYYY-MM-DD and timestamps in UTC.
func toBoarResponse(boar *boardomain.Boar) boarResponse {
	return boarResponse{
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
		CreatedAt: boar.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt: boar.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		CreatedBy: boar.CreatedBy,
		UpdatedBy: boar.UpdatedBy,
	}
}

// toBoarResponses maps a list of domain boars to HTTP response shapes.
// It returns an empty slice instead of null when there are no boars.
func toBoarResponses(boars []*boardomain.Boar) []boarResponse {
	responses := make([]boarResponse, 0, len(boars))
	for _, boar := range boars {
		responses = append(responses, toBoarResponse(boar))
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
	case errors.Is(err, ports.ErrBoarCodeAlreadyUsed):
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
