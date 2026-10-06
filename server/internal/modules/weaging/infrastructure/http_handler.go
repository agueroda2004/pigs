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
	weagingapplication "server/internal/modules/weaging/application"
	weagingdomain "server/internal/modules/weaging/domain"
	"server/internal/modules/weaging/ports"
	platformhttp "server/internal/platform/http"
)

const dateLayout = "2006-01-02"

type CreateWeagingUseCase interface {
	Execute(context.Context, weagingapplication.CreateWeagingCommand) (*weagingdomain.Weaging, error)
}

type ListWeagingsUseCase interface {
	Execute(context.Context, ports.WeagingFilter) ([]*weagingdomain.Weaging, error)
}

type WeagingHandler struct {
	createWeaging   CreateWeagingUseCase
	listWeagings    ListWeagingsUseCase
	authMiddleware  func(http.Handler) http.Handler
	adminMiddleware func(http.Handler) http.Handler
}

// NewWeagingHandler wires the weaging use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewWeagingHandler(
	createWeaging CreateWeagingUseCase,
	listWeagings ListWeagingsUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *WeagingHandler {
	return &WeagingHandler{
		createWeaging:   createWeaging,
		listWeagings:    listWeagings,
		authMiddleware:  authMiddleware,
		adminMiddleware: adminMiddleware,
	}
}

// RegisterRoutes registers the weaging create and list endpoints on the mux.
// Creation is admin-only while the list route only requires authentication.
func (h *WeagingHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/weagings", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/weagings", h.authMiddleware(http.HandlerFunc(h.list)))
}

type createWeagingRequest struct {
	SowID       string   `json:"sow_id"`
	WeagingDate string   `json:"weaging_date"`
	Quantity    int      `json:"quantity"`
	TotalWeight *float64 `json:"total_weight"`
	Destination *string  `json:"destination"`
	Note        *string  `json:"note"`
}

type weagingResponse struct {
	ID          uuid.UUID `json:"id"`
	FarrowingID uuid.UUID `json:"farrowing_id"`
	SowID       uuid.UUID `json:"sow_id"`
	WeagingDate string    `json:"weaging_date"`
	Quantity    int       `json:"quantity"`
	TotalWeight *float64  `json:"total_weight"`
	Destination *string   `json:"destination"`
	Note        *string   `json:"note"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
	CreatedBy   uuid.UUID `json:"created_by"`
	UpdatedBy   uuid.UUID `json:"updated_by"`
}

// create handles POST /api/v1/weagings and registers a weaging.
// It parses the sow, weaging date, quantity, weight, destination and note, reads the
// actor from the context and lets the use case derive the latest farrowing.
func (h *WeagingHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createWeagingRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	sowID, err := uuid.Parse(request.SowID)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La cerda no es válida"))
		return
	}

	weagingDate, err := time.Parse(dateLayout, request.WeagingDate)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La fecha del destete no es válida"))
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	createdWeaging, err := h.createWeaging.Execute(r.Context(), weagingapplication.CreateWeagingCommand{
		SowID:       sowID,
		WeagingDate: weagingDate,
		Quantity:    request.Quantity,
		TotalWeight: request.TotalWeight,
		Destination: request.Destination,
		Note:        request.Note,
		CreatedBy:   actor.UserID,
	})
	if err != nil {
		writeWeagingError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusCreated, toWeagingResponse(createdWeaging))
}

// list handles GET /api/v1/weagings and returns the weagings matching the filter.
// It parses the optional sow_id, from and to query parameters.
func (h *WeagingHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseWeagingFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	weagings, err := h.listWeagings.Execute(r.Context(), filter)
	if err != nil {
		writeWeagingError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toWeagingResponses(weagings))
}

// parseWeagingFilter reads the optional list filters from the query string.
// It returns an error for a malformed sow or date value and rejects a date range
// whose start is later than its end.
func parseWeagingFilter(r *http.Request) (ports.WeagingFilter, error) {
	query := r.URL.Query()
	var filter ports.WeagingFilter

	if sow := strings.TrimSpace(query.Get("sow_id")); sow != "" {
		sowID, err := uuid.Parse(sow)
		if err != nil {
			return ports.WeagingFilter{}, errors.New("El identificador de la cerda no es válido")
		}
		filter.SowID = &sowID
	}

	if from := strings.TrimSpace(query.Get("from")); from != "" {
		fromDate, err := time.Parse(dateLayout, from)
		if err != nil {
			return ports.WeagingFilter{}, errors.New("La fecha inicial no es válida")
		}
		filter.FromDate = &fromDate
	}

	if to := strings.TrimSpace(query.Get("to")); to != "" {
		toDate, err := time.Parse(dateLayout, to)
		if err != nil {
			return ports.WeagingFilter{}, errors.New("La fecha final no es válida")
		}
		filter.ToDate = &toDate
	}

	if filter.FromDate != nil && filter.ToDate != nil && filter.FromDate.After(*filter.ToDate) {
		return ports.WeagingFilter{}, errors.New("La fecha inicial no puede ser posterior a la fecha final")
	}

	return filter, nil
}

// toWeagingResponse maps a domain weaging to the HTTP response shape.
// It formats the weaging date as YYYY-MM-DD and the timestamps in UTC.
func toWeagingResponse(weaging *weagingdomain.Weaging) weagingResponse {
	return weagingResponse{
		ID:          weaging.ID,
		FarrowingID: weaging.FarrowingID,
		SowID:       weaging.SowID,
		WeagingDate: weaging.WeagingDate.UTC().Format(dateLayout),
		Quantity:    weaging.Quantity,
		TotalWeight: weaging.TotalWeight,
		Destination: weaging.Destination,
		Note:        weaging.Note,
		CreatedAt:   weaging.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:   weaging.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		CreatedBy:   weaging.CreatedBy,
		UpdatedBy:   weaging.UpdatedBy,
	}
}

// toWeagingResponses maps a list of domain weagings to HTTP response shapes.
// It returns an empty slice instead of null when there are no weagings.
func toWeagingResponses(weagings []*weagingdomain.Weaging) []weagingResponse {
	responses := make([]weagingResponse, 0, len(weagings))
	for _, weaging := range weagings {
		responses = append(responses, toWeagingResponse(weaging))
	}
	return responses
}

// writeWeagingError maps weaging domain, application and port errors to status codes.
// It falls back to 500 for unrecognized errors.
func writeWeagingError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrWeagingNotFound),
		errors.Is(err, ports.ErrSowNotFound),
		errors.Is(err, ports.ErrFarrowingNotFound):
		status = http.StatusNotFound
	case errors.Is(err, weagingapplication.ErrSowNotLactating),
		errors.Is(err, ports.ErrWeagingAlreadyExists),
		errors.Is(err, farrowingdomain.ErrWeagingQuantityMismatch):
		status = http.StatusConflict
	case errors.Is(err, weagingdomain.ErrInvalidID),
		errors.Is(err, weagingdomain.ErrInvalidFarrowing),
		errors.Is(err, weagingdomain.ErrInvalidSow),
		errors.Is(err, weagingdomain.ErrInvalidWeagingDate),
		errors.Is(err, weagingdomain.ErrInvalidFarrowDate),
		errors.Is(err, weagingdomain.ErrInvalidQuantity),
		errors.Is(err, weagingdomain.ErrInvalidTotalWeight),
		errors.Is(err, weagingdomain.ErrInvalidDestination),
		errors.Is(err, weagingdomain.ErrInvalidNote),
		errors.Is(err, weagingdomain.ErrInvalidCreatedBy),
		errors.Is(err, weagingdomain.ErrInvalidUpdatedBy),
		errors.Is(err, weagingdomain.ErrWeagingDateBeforeFarrowing),
		errors.Is(err, weagingdomain.ErrWeagingDateBeforeEvents),
		errors.Is(err, weagingdomain.ErrWeagingDateInFuture),
		errors.Is(err, farrowingdomain.ErrInvalidPigletQuantity):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
