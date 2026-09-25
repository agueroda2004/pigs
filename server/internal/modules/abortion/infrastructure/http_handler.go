package infrastructure

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	abortionapplication "server/internal/modules/abortion/application"
	abortiondomain "server/internal/modules/abortion/domain"
	"server/internal/modules/abortion/ports"
	authdomain "server/internal/modules/auth/domain"
	platformhttp "server/internal/platform/http"
)

const dateLayout = "2006-01-02"

type CreateAbortionUseCase interface {
	Execute(context.Context, abortionapplication.CreateAbortionCommand) (*abortiondomain.Abortion, error)
}

type ListAbortionsUseCase interface {
	Execute(context.Context, ports.AbortionFilter) ([]*abortiondomain.Abortion, error)
}

type AbortionHandler struct {
	createAbortion  CreateAbortionUseCase
	listAbortions   ListAbortionsUseCase
	authMiddleware  func(http.Handler) http.Handler
	adminMiddleware func(http.Handler) http.Handler
}

// NewAbortionHandler wires the abortion use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewAbortionHandler(
	createAbortion CreateAbortionUseCase,
	listAbortions ListAbortionsUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *AbortionHandler {
	return &AbortionHandler{
		createAbortion:  createAbortion,
		listAbortions:   listAbortions,
		authMiddleware:  authMiddleware,
		adminMiddleware: adminMiddleware,
	}
}

// RegisterRoutes registers the abortion create and list endpoints on the mux.
// Creation is admin-only while the list route only requires authentication.
func (h *AbortionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/abortions", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/abortions", h.authMiddleware(http.HandlerFunc(h.list)))
}

type createAbortionRequest struct {
	SowID        string  `json:"sow_id"`
	AbortionDate string  `json:"abortion_date"`
	Cause        string  `json:"cause"`
	Note         *string `json:"note"`
}

type abortionResponse struct {
	ID           uuid.UUID `json:"id"`
	SowID        uuid.UUID `json:"sow_id"`
	ServiceID    uuid.UUID `json:"service_id"`
	AbortionDate string    `json:"abortion_date"`
	Cause        string    `json:"cause"`
	Note         *string   `json:"note"`
	CreatedAt    string    `json:"created_at"`
	UpdatedAt    string    `json:"updated_at"`
	CreatedBy    uuid.UUID `json:"created_by"`
	UpdatedBy    uuid.UUID `json:"updated_by"`
}

// create handles POST /api/v1/abortions and registers an abortion.
// It parses the date and cause, reads the actor from the context and never
// accepts the service, which is derived by the use case.
func (h *AbortionHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createAbortionRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	sowID, err := uuid.Parse(request.SowID)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La cerda no es válida"))
		return
	}

	abortionDate, err := time.Parse(dateLayout, request.AbortionDate)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La fecha del aborto no es válida"))
		return
	}

	cause, err := abortiondomain.ParseCause(request.Cause)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La causa del aborto no es válida"))
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	createdAbortion, err := h.createAbortion.Execute(r.Context(), abortionapplication.CreateAbortionCommand{
		SowID:        sowID,
		AbortionDate: abortionDate,
		Cause:        cause,
		Note:         request.Note,
		CreatedBy:    actor.UserID,
	})
	if err != nil {
		writeAbortionError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusCreated, toAbortionResponse(createdAbortion))
}

// list handles GET /api/v1/abortions and returns the abortions matching the filter.
// It parses the optional sow_id query parameter.
func (h *AbortionHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseAbortionFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	abortions, err := h.listAbortions.Execute(r.Context(), filter)
	if err != nil {
		writeAbortionError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toAbortionResponses(abortions))
}

// parseAbortionFilter reads the optional list filters from the query string.
// It returns an error for a malformed sow_id value.
func parseAbortionFilter(r *http.Request) (ports.AbortionFilter, error) {
	query := r.URL.Query()
	var filter ports.AbortionFilter

	if sow := strings.TrimSpace(query.Get("sow_id")); sow != "" {
		sowID, err := uuid.Parse(sow)
		if err != nil {
			return ports.AbortionFilter{}, errors.New("El identificador de la cerda no es válido")
		}
		filter.SowID = &sowID
	}

	return filter, nil
}

// toAbortionResponse maps a domain abortion to the HTTP response shape.
// It formats the date as YYYY-MM-DD and the timestamps in UTC.
func toAbortionResponse(abortion *abortiondomain.Abortion) abortionResponse {
	return abortionResponse{
		ID:           abortion.ID,
		SowID:        abortion.SowID,
		ServiceID:    abortion.ServiceID,
		AbortionDate: abortion.AbortionDate.UTC().Format(dateLayout),
		Cause:        string(abortion.Cause),
		Note:         abortion.Note,
		CreatedAt:    abortion.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:    abortion.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		CreatedBy:    abortion.CreatedBy,
		UpdatedBy:    abortion.UpdatedBy,
	}
}

// toAbortionResponses maps a list of domain abortions to HTTP response shapes.
// It returns an empty slice instead of null when there are no abortions.
func toAbortionResponses(abortions []*abortiondomain.Abortion) []abortionResponse {
	responses := make([]abortionResponse, 0, len(abortions))
	for _, abortion := range abortions {
		responses = append(responses, toAbortionResponse(abortion))
	}
	return responses
}

// writeAbortionError maps abortion domain, application and port errors to status codes.
// It falls back to 500 for unrecognized errors.
func writeAbortionError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrAbortionNotFound),
		errors.Is(err, ports.ErrSowNotFound),
		errors.Is(err, ports.ErrServiceNotFound):
		status = http.StatusNotFound
	case errors.Is(err, abortionapplication.ErrSowNotGestating),
		errors.Is(err, abortionapplication.ErrServiceNotConfirmed):
		status = http.StatusConflict
	case errors.Is(err, abortiondomain.ErrInvalidID),
		errors.Is(err, abortiondomain.ErrInvalidSow),
		errors.Is(err, abortiondomain.ErrInvalidService),
		errors.Is(err, abortiondomain.ErrInvalidAbortionDate),
		errors.Is(err, abortiondomain.ErrInvalidLastMount),
		errors.Is(err, abortiondomain.ErrInvalidCause),
		errors.Is(err, abortiondomain.ErrInvalidNote),
		errors.Is(err, abortiondomain.ErrInvalidCreatedBy),
		errors.Is(err, abortiondomain.ErrInvalidUpdatedBy),
		errors.Is(err, abortiondomain.ErrAbortionDateBeforeMount),
		errors.Is(err, abortiondomain.ErrAbortionDateInFuture):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
