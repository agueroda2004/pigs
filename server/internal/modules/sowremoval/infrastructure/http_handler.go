package infrastructure

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	sowremovalapplication "server/internal/modules/sowremoval/application"
	sowremovaldomain "server/internal/modules/sowremoval/domain"
	"server/internal/modules/sowremoval/ports"
	platformhttp "server/internal/platform/http"
)

const dateLayout = "2006-01-02"

type CreateSowRemovalUseCase interface {
	Execute(context.Context, sowremovalapplication.CreateSowRemovalCommand) (*sowremovaldomain.SowRemoval, error)
}

type ListSowRemovalsUseCase interface {
	Execute(context.Context, ports.SowRemovalFilter) ([]*sowremovaldomain.SowRemoval, error)
}

type SowRemovalHandler struct {
	createSowRemoval CreateSowRemovalUseCase
	listSowRemovals  ListSowRemovalsUseCase
	authMiddleware   func(http.Handler) http.Handler
	adminMiddleware  func(http.Handler) http.Handler
}

// NewSowRemovalHandler wires the removal use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewSowRemovalHandler(
	createSowRemoval CreateSowRemovalUseCase,
	listSowRemovals ListSowRemovalsUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *SowRemovalHandler {
	return &SowRemovalHandler{
		createSowRemoval: createSowRemoval,
		listSowRemovals:  listSowRemovals,
		authMiddleware:   authMiddleware,
		adminMiddleware:  adminMiddleware,
	}
}

// RegisterRoutes registers the removal create and list endpoints on the mux.
// Creation is admin-only while the list route only requires authentication.
func (h *SowRemovalHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/sow-removals", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/sow-removals", h.authMiddleware(http.HandlerFunc(h.list)))
}

type createSowRemovalRequest struct {
	SowID       string  `json:"sow_id"`
	RemovalDate string  `json:"removal_date"`
	Type        string  `json:"type"`
	Reason      string  `json:"reason"`
	Note        *string `json:"note"`
}

type sowRemovalResponse struct {
	ID          uuid.UUID `json:"id"`
	SowID       uuid.UUID `json:"sow_id"`
	RemovalDate string    `json:"removal_date"`
	Type        string    `json:"type"`
	Reason      string    `json:"reason"`
	Note        *string   `json:"note"`
	LastState   string    `json:"last_state"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
	CreatedBy   uuid.UUID `json:"created_by"`
	UpdatedBy   uuid.UUID `json:"updated_by"`
}

// create handles POST /api/v1/sow-removals and registers a removal.
// It parses the date, type and reason, reads the actor from the context and
// never accepts the last state, which is derived by the use case.
func (h *SowRemovalHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createSowRemovalRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	sowID, err := uuid.Parse(request.SowID)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La cerda no es válida"))
		return
	}

	removalDate, err := time.Parse(dateLayout, request.RemovalDate)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La fecha de la baja no es válida"))
		return
	}

	removalType, err := sowremovaldomain.ParseType(request.Type)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El tipo de baja no es válido"))
		return
	}

	reason, err := sowremovaldomain.ParseReason(request.Reason)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El motivo de baja no es válido"))
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	createdRemoval, err := h.createSowRemoval.Execute(r.Context(), sowremovalapplication.CreateSowRemovalCommand{
		SowID:       sowID,
		RemovalDate: removalDate,
		Type:        removalType,
		Reason:      reason,
		Note:        request.Note,
		CreatedBy:   actor.UserID,
	})
	if err != nil {
		writeSowRemovalError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusCreated, toSowRemovalResponse(createdRemoval))
}

// list handles GET /api/v1/sow-removals and returns the removals matching the filter.
// It parses the optional sow_id query parameter.
func (h *SowRemovalHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseSowRemovalFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	removals, err := h.listSowRemovals.Execute(r.Context(), filter)
	if err != nil {
		writeSowRemovalError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toSowRemovalResponses(removals))
}

// parseSowRemovalFilter reads the optional list filters from the query string.
// It returns an error for a malformed sow_id value.
func parseSowRemovalFilter(r *http.Request) (ports.SowRemovalFilter, error) {
	query := r.URL.Query()
	var filter ports.SowRemovalFilter

	if sow := strings.TrimSpace(query.Get("sow_id")); sow != "" {
		sowID, err := uuid.Parse(sow)
		if err != nil {
			return ports.SowRemovalFilter{}, errors.New("El identificador de la cerda no es válido")
		}
		filter.SowID = &sowID
	}

	return filter, nil
}

// toSowRemovalResponse maps a domain removal to the HTTP response shape.
// It formats the date as YYYY-MM-DD and the timestamps in UTC.
func toSowRemovalResponse(removal *sowremovaldomain.SowRemoval) sowRemovalResponse {
	return sowRemovalResponse{
		ID:          removal.ID,
		SowID:       removal.SowID,
		RemovalDate: removal.RemovalDate.UTC().Format(dateLayout),
		Type:        string(removal.Type),
		Reason:      string(removal.Reason),
		Note:        removal.Note,
		LastState:   removal.LastState,
		CreatedAt:   removal.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:   removal.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		CreatedBy:   removal.CreatedBy,
		UpdatedBy:   removal.UpdatedBy,
	}
}

// toSowRemovalResponses maps a list of domain removals to HTTP response shapes.
// It returns an empty slice instead of null when there are no removals.
func toSowRemovalResponses(removals []*sowremovaldomain.SowRemoval) []sowRemovalResponse {
	responses := make([]sowRemovalResponse, 0, len(removals))
	for _, removal := range removals {
		responses = append(responses, toSowRemovalResponse(removal))
	}
	return responses
}

// writeSowRemovalError maps removal domain, application and port errors to status codes.
// It falls back to 500 for unrecognized errors.
func writeSowRemovalError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrSowRemovalNotFound),
		errors.Is(err, ports.ErrSowNotFound),
		errors.Is(err, ports.ErrServiceNotFound),
		errors.Is(err, ports.ErrAbortionNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ports.ErrSowAlreadyRemoved),
		errors.Is(err, sowremovalapplication.ErrSowNotRemovable):
		status = http.StatusConflict
	case errors.Is(err, sowremovaldomain.ErrInvalidID),
		errors.Is(err, sowremovaldomain.ErrInvalidSow),
		errors.Is(err, sowremovaldomain.ErrInvalidRemovalDate),
		errors.Is(err, sowremovaldomain.ErrRemovalDateInFuture),
		errors.Is(err, sowremovaldomain.ErrInvalidType),
		errors.Is(err, sowremovaldomain.ErrInvalidReason),
		errors.Is(err, sowremovaldomain.ErrInvalidNote),
		errors.Is(err, sowremovaldomain.ErrInvalidCreatedBy),
		errors.Is(err, sowremovaldomain.ErrInvalidUpdatedBy),
		errors.Is(err, sowremovaldomain.ErrSowNotRemovable),
		errors.Is(err, sowremovaldomain.ErrRemovalDateBeforeWeaning),
		errors.Is(err, sowremovaldomain.ErrRemovalDateBeforeService),
		errors.Is(err, sowremovaldomain.ErrRemovalDateBeforeAbortion):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
