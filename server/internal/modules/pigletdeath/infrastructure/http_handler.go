package infrastructure

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	farrowingdomain "server/internal/modules/farrowing/domain"
	pigletdeathapplication "server/internal/modules/pigletdeath/application"
	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	"server/internal/modules/pigletdeath/ports"
	platformhttp "server/internal/platform/http"
)

const dateLayout = "2006-01-02"

type CreatePigletDeathUseCase interface {
	Execute(context.Context, pigletdeathapplication.CreatePigletDeathCommand) (*pigletdeathdomain.PigletDeath, error)
}

type ListPigletDeathsUseCase interface {
	Execute(context.Context, ports.PigletDeathFilter) ([]*pigletdeathdomain.PigletDeath, error)
}

type PigletDeathHandler struct {
	createPigletDeath CreatePigletDeathUseCase
	listPigletDeaths  ListPigletDeathsUseCase
	authMiddleware    func(http.Handler) http.Handler
	adminMiddleware   func(http.Handler) http.Handler
}

// NewPigletDeathHandler wires the piglet death use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewPigletDeathHandler(
	createPigletDeath CreatePigletDeathUseCase,
	listPigletDeaths ListPigletDeathsUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *PigletDeathHandler {
	return &PigletDeathHandler{
		createPigletDeath: createPigletDeath,
		listPigletDeaths:  listPigletDeaths,
		authMiddleware:    authMiddleware,
		adminMiddleware:   adminMiddleware,
	}
}

// RegisterRoutes registers the piglet death create and list endpoints on the mux.
// Creation is admin-only while the list route only requires authentication.
func (h *PigletDeathHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/piglet-deaths", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/piglet-deaths", h.authMiddleware(http.HandlerFunc(h.list)))
}

type createPigletDeathRequest struct {
	SowID      string   `json:"sow_id"`
	OperatorID string   `json:"operator_id"`
	DeathDate  string   `json:"death_date"`
	Quantity   int      `json:"quantity"`
	Weight     *float64 `json:"weight"`
	Cause      string   `json:"cause"`
	Turn       string   `json:"turn"`
	Note       *string  `json:"note"`
}

type pigletDeathResponse struct {
	ID           uuid.UUID `json:"id"`
	FarrowingID  uuid.UUID `json:"farrowing_id"`
	SowID        uuid.UUID `json:"sow_id"`
	OperatorID   uuid.UUID `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	DeathDate    string    `json:"death_date"`
	Quantity     int       `json:"quantity"`
	Weight       *float64  `json:"weight"`
	Cause        string    `json:"cause"`
	Turn         string    `json:"turn"`
	Note         *string   `json:"note"`
	CreatedAt    string    `json:"created_at"`
	UpdatedAt    string    `json:"updated_at"`
	CreatedBy    uuid.UUID `json:"created_by"`
	UpdatedBy    uuid.UUID `json:"updated_by"`
}

// create handles POST /api/v1/piglet-deaths and registers a piglet death.
// It parses the sow, operator, date, quantity, cause and turn, reads the actor
// from the context and lets the use case resolve the sow's latest farrowing.
func (h *PigletDeathHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createPigletDeathRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	sowID, err := uuid.Parse(request.SowID)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La cerda no es válida"))
		return
	}

	operatorID, err := uuid.Parse(request.OperatorID)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El operador no es válido"))
		return
	}

	deathDate, err := time.Parse(dateLayout, request.DeathDate)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La fecha de la muerte no es válida"))
		return
	}

	cause, err := pigletdeathdomain.ParseCause(request.Cause)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La causa de la muerte no es válida"))
		return
	}

	turn, err := pigletdeathdomain.ParseTurn(request.Turn)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El turno no es válido"))
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	createdDeath, err := h.createPigletDeath.Execute(r.Context(), pigletdeathapplication.CreatePigletDeathCommand{
		SowID:      sowID,
		OperatorID: operatorID,
		DeathDate:  deathDate,
		Quantity:   request.Quantity,
		Weight:     request.Weight,
		Cause:      cause,
		Turn:       turn,
		Note:       request.Note,
		CreatedBy:  actor.UserID,
	})
	if err != nil {
		writePigletDeathError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusCreated, toPigletDeathResponse(createdDeath))
}

// list handles GET /api/v1/piglet-deaths and returns the deaths matching the filter.
// It parses the optional sow_id, from and to query parameters.
func (h *PigletDeathHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parsePigletDeathFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	deaths, err := h.listPigletDeaths.Execute(r.Context(), filter)
	if err != nil {
		writePigletDeathError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toPigletDeathResponses(deaths))
}

// parsePigletDeathFilter reads the optional list filters from the query string.
// It returns an error for a malformed sow_id, from or to value and rejects a
// date range whose start is later than its end.
func parsePigletDeathFilter(r *http.Request) (ports.PigletDeathFilter, error) {
	query := r.URL.Query()
	var filter ports.PigletDeathFilter

	if sow := strings.TrimSpace(query.Get("sow_id")); sow != "" {
		sowID, err := uuid.Parse(sow)
		if err != nil {
			return ports.PigletDeathFilter{}, errors.New("El identificador de la cerda no es válido")
		}
		filter.SowID = &sowID
	}

	if from := strings.TrimSpace(query.Get("from")); from != "" {
		fromDate, err := time.Parse(dateLayout, from)
		if err != nil {
			return ports.PigletDeathFilter{}, errors.New("La fecha inicial no es válida")
		}
		filter.FromDate = &fromDate
	}

	if to := strings.TrimSpace(query.Get("to")); to != "" {
		toDate, err := time.Parse(dateLayout, to)
		if err != nil {
			return ports.PigletDeathFilter{}, errors.New("La fecha final no es válida")
		}
		filter.ToDate = &toDate
	}

	if filter.FromDate != nil && filter.ToDate != nil && filter.FromDate.After(*filter.ToDate) {
		return ports.PigletDeathFilter{}, errors.New("La fecha inicial no puede ser posterior a la fecha final")
	}

	return filter, nil
}

// toPigletDeathResponse maps a domain piglet death to the HTTP response shape.
// It formats the date as YYYY-MM-DD and the timestamps in UTC.
func toPigletDeathResponse(death *pigletdeathdomain.PigletDeath) pigletDeathResponse {
	return pigletDeathResponse{
		ID:           death.ID,
		FarrowingID:  death.FarrowingID,
		SowID:        death.SowID,
		OperatorID:   death.OperatorID,
		OperatorName: death.OperatorName,
		DeathDate:    death.DeathDate.UTC().Format(dateLayout),
		Quantity:     death.Quantity,
		Weight:       death.Weight,
		Cause:        string(death.Cause),
		Turn:         string(death.Turn),
		Note:         death.Note,
		CreatedAt:    death.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:    death.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		CreatedBy:    death.CreatedBy,
		UpdatedBy:    death.UpdatedBy,
	}
}

// toPigletDeathResponses maps a list of domain piglet deaths to HTTP response shapes.
// It returns an empty slice instead of null when there are no deaths.
func toPigletDeathResponses(deaths []*pigletdeathdomain.PigletDeath) []pigletDeathResponse {
	responses := make([]pigletDeathResponse, 0, len(deaths))
	for _, death := range deaths {
		responses = append(responses, toPigletDeathResponse(death))
	}
	return responses
}

// writePigletDeathError maps piglet death domain, application and port errors to status codes.
// It falls back to 500 for unrecognized errors.
func writePigletDeathError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrPigletDeathNotFound),
		errors.Is(err, ports.ErrSowNotFound),
		errors.Is(err, ports.ErrFarrowingNotFound),
		errors.Is(err, ports.ErrOperatorNotFound):
		status = http.StatusNotFound
	case errors.Is(err, pigletdeathapplication.ErrSowNotLactating),
		errors.Is(err, farrowingdomain.ErrInsufficientPiglets):
		status = http.StatusConflict
	case errors.Is(err, pigletdeathdomain.ErrInvalidID),
		errors.Is(err, pigletdeathdomain.ErrInvalidFarrowing),
		errors.Is(err, pigletdeathdomain.ErrInvalidSow),
		errors.Is(err, pigletdeathdomain.ErrInvalidOperator),
		errors.Is(err, pigletdeathdomain.ErrInvalidDeathDate),
		errors.Is(err, pigletdeathdomain.ErrInvalidFarrowDate),
		errors.Is(err, pigletdeathdomain.ErrInvalidQuantity),
		errors.Is(err, pigletdeathdomain.ErrInvalidWeight),
		errors.Is(err, pigletdeathdomain.ErrInvalidCause),
		errors.Is(err, pigletdeathdomain.ErrInvalidTurn),
		errors.Is(err, pigletdeathdomain.ErrInvalidNote),
		errors.Is(err, pigletdeathdomain.ErrInvalidCreatedBy),
		errors.Is(err, pigletdeathdomain.ErrInvalidUpdatedBy),
		errors.Is(err, pigletdeathdomain.ErrDeathDateBeforeFarrowing),
		errors.Is(err, pigletdeathdomain.ErrDeathDateInFuture),
		errors.Is(err, farrowingdomain.ErrInvalidPigletQuantity):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
