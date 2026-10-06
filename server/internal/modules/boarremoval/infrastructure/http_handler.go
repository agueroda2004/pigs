package infrastructure

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	boarremovalapplication "server/internal/modules/boarremoval/application"
	boarremovaldomain "server/internal/modules/boarremoval/domain"
	"server/internal/modules/boarremoval/ports"
	platformhttp "server/internal/platform/http"
)

const dateLayout = "2006-01-02"

type CreateBoarRemovalUseCase interface {
	Execute(context.Context, boarremovalapplication.CreateBoarRemovalCommand) (*boarremovaldomain.BoarRemoval, error)
}

type UpdateBoarRemovalUseCase interface {
	Execute(context.Context, uuid.UUID, boarremovalapplication.UpdateBoarRemovalCommand) (*boarremovaldomain.BoarRemoval, error)
}

type DeleteBoarRemovalUseCase interface {
	Execute(context.Context, boarremovalapplication.DeleteBoarRemovalCommand) error
}

type ListBoarRemovalsUseCase interface {
	Execute(context.Context, ports.BoarRemovalFilter) ([]*boarremovaldomain.BoarRemoval, error)
}

type BoarRemovalHandler struct {
	createBoarRemoval CreateBoarRemovalUseCase
	updateBoarRemoval UpdateBoarRemovalUseCase
	deleteBoarRemoval DeleteBoarRemovalUseCase
	listBoarRemovals  ListBoarRemovalsUseCase
	authMiddleware    func(http.Handler) http.Handler
	adminMiddleware   func(http.Handler) http.Handler
}

// NewBoarRemovalHandler wires the removal use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewBoarRemovalHandler(
	createBoarRemoval CreateBoarRemovalUseCase,
	updateBoarRemoval UpdateBoarRemovalUseCase,
	deleteBoarRemoval DeleteBoarRemovalUseCase,
	listBoarRemovals ListBoarRemovalsUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *BoarRemovalHandler {
	return &BoarRemovalHandler{
		createBoarRemoval: createBoarRemoval,
		updateBoarRemoval: updateBoarRemoval,
		deleteBoarRemoval: deleteBoarRemoval,
		listBoarRemovals:  listBoarRemovals,
		authMiddleware:    authMiddleware,
		adminMiddleware:   adminMiddleware,
	}
}

// RegisterRoutes registers the removal create, list, update and delete endpoints.
// Write routes are admin-only while the list route only requires authentication.
func (h *BoarRemovalHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/boar-removals", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/boar-removals", h.authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("PATCH /api/v1/boar-removals/{id}", h.adminMiddleware(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /api/v1/boar-removals/{id}", h.adminMiddleware(http.HandlerFunc(h.delete)))
}

type createBoarRemovalRequest struct {
	BoarID      string  `json:"boar_id"`
	RemovalDate string  `json:"removal_date"`
	Type        string  `json:"type"`
	Reason      string  `json:"reason"`
	Note        *string `json:"note"`
}

type updateBoarRemovalRequest struct {
	RemovalDate *string `json:"removal_date"`
	Type        *string `json:"type"`
	Reason      *string `json:"reason"`
	Note        *string `json:"note"`
}

type boarRemovalResponse struct {
	ID          uuid.UUID `json:"id"`
	BoarID      uuid.UUID `json:"boar_id"`
	RemovalDate string    `json:"removal_date"`
	Type        string    `json:"type"`
	Reason      string    `json:"reason"`
	Note        *string   `json:"note"`
	LastState   string    `json:"last_state"`
}

// create handles POST /api/v1/boar-removals and registers a removal.
// It parses the date, type and reason, reads the actor from the context and
// never accepts the last state, which is derived by the use case.
func (h *BoarRemovalHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createBoarRemovalRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	boarID, err := uuid.Parse(request.BoarID)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El verraco no es válido"))
		return
	}

	removalDate, err := time.Parse(dateLayout, request.RemovalDate)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La fecha de la baja no es válida"))
		return
	}

	removalType, err := boarremovaldomain.ParseType(request.Type)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El tipo de baja no es válido"))
		return
	}

	reason, err := boarremovaldomain.ParseReason(request.Reason)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El motivo de baja no es válido"))
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if _, err := h.createBoarRemoval.Execute(r.Context(), boarremovalapplication.CreateBoarRemovalCommand{
		BoarID:      boarID,
		RemovalDate: removalDate,
		Type:        removalType,
		Reason:      reason,
		Note:        request.Note,
		CreatedBy:   actor.UserID,
	}); err != nil {
		writeBoarRemovalError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// update handles PATCH /api/v1/boar-removals/{id} and applies the provided fields.
// It parses the id path value, reads the actor and never accepts the boar, which
// is immutable for an existing removal.
func (h *BoarRemovalHandler) update(w http.ResponseWriter, r *http.Request) {
	removalID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador de la baja no es válido"))
		return
	}

	var request updateBoarRemovalRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	removalDate, err := parseBoarRemovalDate(request.RemovalDate)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La fecha de la baja no es válida"))
		return
	}

	var removalType *boarremovaldomain.Type
	if request.Type != nil {
		parsedType, err := boarremovaldomain.ParseType(*request.Type)
		if err != nil {
			platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El tipo de baja no es válido"))
			return
		}
		removalType = &parsedType
	}

	var reason *boarremovaldomain.Reason
	if request.Reason != nil {
		parsedReason, err := boarremovaldomain.ParseReason(*request.Reason)
		if err != nil {
			platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El motivo de baja no es válido"))
			return
		}
		reason = &parsedReason
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if _, err := h.updateBoarRemoval.Execute(r.Context(), removalID, boarremovalapplication.UpdateBoarRemovalCommand{
		RemovalDate: removalDate,
		Type:        removalType,
		Reason:      reason,
		Note:        request.Note,
		UpdatedBy:   actor.UserID,
	}); err != nil {
		writeBoarRemovalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// parseBoarRemovalDate parses an optional removal date using the shared layout.
// It returns nil when value is nil and an error for an unparseable date.
func parseBoarRemovalDate(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := time.Parse(dateLayout, *value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// delete handles DELETE /api/v1/boar-removals/{id} and undoes the removal.
// It parses the id path value, reads the actor and returns no content.
func (h *BoarRemovalHandler) delete(w http.ResponseWriter, r *http.Request) {
	removalID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador de la baja no es válido"))
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if err := h.deleteBoarRemoval.Execute(r.Context(), boarremovalapplication.DeleteBoarRemovalCommand{
		ID:        removalID,
		DeletedBy: actor.UserID,
	}); err != nil {
		writeBoarRemovalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// list handles GET /api/v1/boar-removals and returns the removals matching the filter.
// It parses the optional boar_id query parameter.
func (h *BoarRemovalHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseBoarRemovalFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	removals, err := h.listBoarRemovals.Execute(r.Context(), filter)
	if err != nil {
		writeBoarRemovalError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toBoarRemovalResponses(removals))
}

// parseBoarRemovalFilter reads the optional list filters from the query string.
// It returns an error for a malformed boar_id value.
func parseBoarRemovalFilter(r *http.Request) (ports.BoarRemovalFilter, error) {
	query := r.URL.Query()
	var filter ports.BoarRemovalFilter

	if boar := strings.TrimSpace(query.Get("boar_id")); boar != "" {
		boarID, err := uuid.Parse(boar)
		if err != nil {
			return ports.BoarRemovalFilter{}, errors.New("El identificador del verraco no es válido")
		}
		filter.BoarID = &boarID
	}

	return filter, nil
}

// toBoarRemovalResponse maps a domain removal to the HTTP response shape.
// It formats the date as YYYY-MM-DD and omits the audit fields.
func toBoarRemovalResponse(removal *boarremovaldomain.BoarRemoval) boarRemovalResponse {
	return boarRemovalResponse{
		ID:          removal.ID,
		BoarID:      removal.BoarID,
		RemovalDate: removal.RemovalDate.UTC().Format(dateLayout),
		Type:        string(removal.Type),
		Reason:      string(removal.Reason),
		Note:        removal.Note,
		LastState:   removal.LastState,
	}
}

// toBoarRemovalResponses maps a list of domain removals to HTTP response shapes.
// It returns an empty slice instead of null when there are no removals.
func toBoarRemovalResponses(removals []*boarremovaldomain.BoarRemoval) []boarRemovalResponse {
	responses := make([]boarRemovalResponse, 0, len(removals))
	for _, removal := range removals {
		responses = append(responses, toBoarRemovalResponse(removal))
	}
	return responses
}

// writeBoarRemovalError maps removal domain, application and port errors to status codes.
// It falls back to 500 for unrecognized errors.
func writeBoarRemovalError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrBoarRemovalNotFound),
		errors.Is(err, ports.ErrBoarNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ports.ErrBoarAlreadyRemoved),
		errors.Is(err, boarremovalapplication.ErrBoarNotRemovable),
		errors.Is(err, boarremovalapplication.ErrBoarStateMismatch):
		status = http.StatusConflict
	case errors.Is(err, boarremovaldomain.ErrInvalidID),
		errors.Is(err, boarremovaldomain.ErrInvalidBoar),
		errors.Is(err, boarremovaldomain.ErrInvalidRemovalDate),
		errors.Is(err, boarremovaldomain.ErrRemovalDateInFuture),
		errors.Is(err, boarremovaldomain.ErrInvalidType),
		errors.Is(err, boarremovaldomain.ErrInvalidReason),
		errors.Is(err, boarremovaldomain.ErrInvalidNote),
		errors.Is(err, boarremovaldomain.ErrInvalidUpdate),
		errors.Is(err, boarremovaldomain.ErrInvalidCreatedBy),
		errors.Is(err, boarremovaldomain.ErrInvalidUpdatedBy),
		errors.Is(err, boarremovaldomain.ErrBoarNotRemovable),
		errors.Is(err, boarremovaldomain.ErrRemovalDateBeforeEntry),
		errors.Is(err, boarremovaldomain.ErrRemovalDateBeforeMount):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
