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
	partialweagingapplication "server/internal/modules/partialweaging/application"
	partialweagingdomain "server/internal/modules/partialweaging/domain"
	"server/internal/modules/partialweaging/ports"
	platformhttp "server/internal/platform/http"
)

const dateLayout = "2006-01-02"

type CreatePartialWeagingUseCase interface {
	Execute(context.Context, partialweagingapplication.CreatePartialWeagingCommand) (*partialweagingdomain.PartialWeaging, error)
}

type ListPartialWeagingsUseCase interface {
	Execute(context.Context, ports.PartialWeagingFilter) ([]*partialweagingdomain.PartialWeaging, error)
}

type PartialWeagingHandler struct {
	createPartialWeaging CreatePartialWeagingUseCase
	listPartialWeagings  ListPartialWeagingsUseCase
	authMiddleware       func(http.Handler) http.Handler
	adminMiddleware      func(http.Handler) http.Handler
}

// NewPartialWeagingHandler wires the partial weaging use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewPartialWeagingHandler(
	createPartialWeaging CreatePartialWeagingUseCase,
	listPartialWeagings ListPartialWeagingsUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *PartialWeagingHandler {
	return &PartialWeagingHandler{
		createPartialWeaging: createPartialWeaging,
		listPartialWeagings:  listPartialWeagings,
		authMiddleware:       authMiddleware,
		adminMiddleware:      adminMiddleware,
	}
}

// RegisterRoutes registers the partial weaging create and list endpoints on the mux.
// Creation is admin-only while the list route only requires authentication.
func (h *PartialWeagingHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/partial-weagings", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/partial-weagings", h.authMiddleware(http.HandlerFunc(h.list)))
}

type createPartialWeagingRequest struct {
	SowID       string   `json:"sow_id"`
	WeagingDate string   `json:"weaging_date"`
	Quantity    int      `json:"quantity"`
	TotalWeight *float64 `json:"total_weight"`
	Type        string   `json:"type"`
	Note        *string  `json:"note"`
}

type partialWeagingResponse struct {
	ID          uuid.UUID `json:"id"`
	FarrowingID uuid.UUID `json:"farrowing_id"`
	SowID       uuid.UUID `json:"sow_id"`
	WeagingDate string    `json:"weaging_date"`
	Quantity    int       `json:"quantity"`
	TotalWeight *float64  `json:"total_weight"`
	Type        string    `json:"type"`
	Note        *string   `json:"note"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
	CreatedBy   uuid.UUID `json:"created_by"`
	UpdatedBy   uuid.UUID `json:"updated_by"`
}

// create handles POST /api/v1/partial-weagings and registers a partial weaging.
// It parses the sow, weaging date, quantity, weight, type and note, reads the actor
// from the context and lets the use case derive the latest farrowing.
func (h *PartialWeagingHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createPartialWeagingRequest
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
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La fecha del destete parcial no es válida"))
		return
	}

	weagingType, err := partialweagingdomain.ParseType(request.Type)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	createdWeaging, err := h.createPartialWeaging.Execute(r.Context(), partialweagingapplication.CreatePartialWeagingCommand{
		SowID:       sowID,
		WeagingDate: weagingDate,
		Quantity:    request.Quantity,
		TotalWeight: request.TotalWeight,
		Type:        weagingType,
		Note:        request.Note,
		CreatedBy:   actor.UserID,
	})
	if err != nil {
		writePartialWeagingError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusCreated, toPartialWeagingResponse(createdWeaging))
}

// list handles GET /api/v1/partial-weagings and returns the partial weagings matching the filter.
// It parses the optional sow_id, from and to query parameters.
func (h *PartialWeagingHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parsePartialWeagingFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	weagings, err := h.listPartialWeagings.Execute(r.Context(), filter)
	if err != nil {
		writePartialWeagingError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toPartialWeagingResponses(weagings))
}

// parsePartialWeagingFilter reads the optional list filters from the query string.
// It returns an error for a malformed sow or date value and rejects a date range
// whose start is later than its end.
func parsePartialWeagingFilter(r *http.Request) (ports.PartialWeagingFilter, error) {
	query := r.URL.Query()
	var filter ports.PartialWeagingFilter

	if sow := strings.TrimSpace(query.Get("sow_id")); sow != "" {
		sowID, err := uuid.Parse(sow)
		if err != nil {
			return ports.PartialWeagingFilter{}, errors.New("El identificador de la cerda no es válido")
		}
		filter.SowID = &sowID
	}

	if from := strings.TrimSpace(query.Get("from")); from != "" {
		fromDate, err := time.Parse(dateLayout, from)
		if err != nil {
			return ports.PartialWeagingFilter{}, errors.New("La fecha inicial no es válida")
		}
		filter.FromDate = &fromDate
	}

	if to := strings.TrimSpace(query.Get("to")); to != "" {
		toDate, err := time.Parse(dateLayout, to)
		if err != nil {
			return ports.PartialWeagingFilter{}, errors.New("La fecha final no es válida")
		}
		filter.ToDate = &toDate
	}

	if filter.FromDate != nil && filter.ToDate != nil && filter.FromDate.After(*filter.ToDate) {
		return ports.PartialWeagingFilter{}, errors.New("La fecha inicial no puede ser posterior a la fecha final")
	}

	return filter, nil
}

// toPartialWeagingResponse maps a domain partial weaging to the HTTP response shape.
// It formats the weaging date as YYYY-MM-DD and the timestamps in UTC.
func toPartialWeagingResponse(weaging *partialweagingdomain.PartialWeaging) partialWeagingResponse {
	return partialWeagingResponse{
		ID:          weaging.ID,
		FarrowingID: weaging.FarrowingID,
		SowID:       weaging.SowID,
		WeagingDate: weaging.WeagingDate.UTC().Format(dateLayout),
		Quantity:    weaging.Quantity,
		TotalWeight: weaging.TotalWeight,
		Type:        string(weaging.Type),
		Note:        weaging.Note,
		CreatedAt:   weaging.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:   weaging.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		CreatedBy:   weaging.CreatedBy,
		UpdatedBy:   weaging.UpdatedBy,
	}
}

// toPartialWeagingResponses maps a list of domain partial weagings to HTTP response shapes.
// It returns an empty slice instead of null when there are no partial weagings.
func toPartialWeagingResponses(weagings []*partialweagingdomain.PartialWeaging) []partialWeagingResponse {
	responses := make([]partialWeagingResponse, 0, len(weagings))
	for _, weaging := range weagings {
		responses = append(responses, toPartialWeagingResponse(weaging))
	}
	return responses
}

// writePartialWeagingError maps partial weaging domain, application and port errors to status codes.
// It falls back to 500 for unrecognized errors.
func writePartialWeagingError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrPartialWeagingNotFound),
		errors.Is(err, ports.ErrSowNotFound),
		errors.Is(err, ports.ErrFarrowingNotFound):
		status = http.StatusNotFound
	case errors.Is(err, partialweagingapplication.ErrSowNotLactating),
		errors.Is(err, farrowingdomain.ErrInsufficientPiglets):
		status = http.StatusConflict
	case errors.Is(err, partialweagingdomain.ErrInvalidID),
		errors.Is(err, partialweagingdomain.ErrInvalidFarrowing),
		errors.Is(err, partialweagingdomain.ErrInvalidSow),
		errors.Is(err, partialweagingdomain.ErrInvalidWeagingDate),
		errors.Is(err, partialweagingdomain.ErrInvalidFarrowDate),
		errors.Is(err, partialweagingdomain.ErrInvalidQuantity),
		errors.Is(err, partialweagingdomain.ErrInvalidTotalWeight),
		errors.Is(err, partialweagingdomain.ErrInvalidType),
		errors.Is(err, partialweagingdomain.ErrInvalidNote),
		errors.Is(err, partialweagingdomain.ErrInvalidCreatedBy),
		errors.Is(err, partialweagingdomain.ErrInvalidUpdatedBy),
		errors.Is(err, partialweagingdomain.ErrWeagingDateBeforeFarrowing),
		errors.Is(err, partialweagingdomain.ErrWeagingDateBeforeEvents),
		errors.Is(err, partialweagingdomain.ErrWeagingDateInFuture),
		errors.Is(err, farrowingdomain.ErrInvalidPigletQuantity):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
