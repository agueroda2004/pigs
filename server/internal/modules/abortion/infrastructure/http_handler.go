package infrastructure

import (
	"context"
	"errors"
	"net/http"
	"strconv"
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

// maxSowCodeLength mirrors the length of the sows.code column.
// A longer sow code filter is rejected as invalid.
const maxSowCodeLength = 50

type CreateAbortionUseCase interface {
	Execute(context.Context, abortionapplication.CreateAbortionCommand) (*abortiondomain.Abortion, error)
}

type ListAbortionsUseCase interface {
	Execute(context.Context, ports.AbortionFilter, int) (abortionapplication.AbortionPage, error)
}

type UpdateAbortionUseCase interface {
	Execute(context.Context, uuid.UUID, abortionapplication.UpdateAbortionCommand) error
}

type DeleteAbortionUseCase interface {
	Execute(context.Context, abortionapplication.DeleteAbortionCommand) error
}

type AbortionHandler struct {
	createAbortion  CreateAbortionUseCase
	listAbortions   ListAbortionsUseCase
	updateAbortion  UpdateAbortionUseCase
	deleteAbortion  DeleteAbortionUseCase
	authMiddleware  func(http.Handler) http.Handler
	adminMiddleware func(http.Handler) http.Handler
}

// NewAbortionHandler wires the abortion use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewAbortionHandler(
	createAbortion CreateAbortionUseCase,
	listAbortions ListAbortionsUseCase,
	updateAbortion UpdateAbortionUseCase,
	deleteAbortion DeleteAbortionUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *AbortionHandler {
	return &AbortionHandler{
		createAbortion:  createAbortion,
		listAbortions:   listAbortions,
		updateAbortion:  updateAbortion,
		deleteAbortion:  deleteAbortion,
		authMiddleware:  authMiddleware,
		adminMiddleware: adminMiddleware,
	}
}

// RegisterRoutes registers the abortion create, list, update and delete endpoints.
// Write routes are admin-only while the list route only requires authentication.
func (h *AbortionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/abortions", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/abortions", h.authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("PATCH /api/v1/abortions/{id}", h.adminMiddleware(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /api/v1/abortions/{id}", h.adminMiddleware(http.HandlerFunc(h.delete)))
}

type createAbortionRequest struct {
	SowID        string  `json:"sow_id"`
	AbortionDate string  `json:"abortion_date"`
	Cause        string  `json:"cause"`
	Note         *string `json:"note"`
}

type updateAbortionRequest struct {
	AbortionDate *string `json:"abortion_date"`
	Cause        *string `json:"cause"`
	Note         *string `json:"note"`
}

// abortionResponse is the abortion list shape without the audit fields.
type abortionResponse struct {
	ID           uuid.UUID `json:"id"`
	SowID        uuid.UUID `json:"sow_id"`
	SowCode      string    `json:"sow_code"`
	ServiceID    uuid.UUID `json:"service_id"`
	AbortionDate string    `json:"abortion_date"`
	Cause        string    `json:"cause"`
	Note         *string   `json:"note"`
}

// abortionPageResponse is one page of abortions together with its pagination metadata.
type abortionPageResponse struct {
	Items      []abortionResponse `json:"items"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
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

	if _, err := h.createAbortion.Execute(r.Context(), abortionapplication.CreateAbortionCommand{
		SowID:        sowID,
		AbortionDate: abortionDate,
		Cause:        cause,
		Note:         request.Note,
		CreatedBy:    actor.UserID,
	}); err != nil {
		writeAbortionError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// update handles PATCH /api/v1/abortions/{id} and applies the editable fields.
// It parses the id path value and never accepts the sow or the service, which
// stay fixed, then returns 200 without a body.
func (h *AbortionHandler) update(w http.ResponseWriter, r *http.Request) {
	abortionID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del aborto no es válido"))
		return
	}

	var request updateAbortionRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	var abortionDate *time.Time
	if request.AbortionDate != nil {
		parsed, err := time.Parse(dateLayout, *request.AbortionDate)
		if err != nil {
			platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La fecha del aborto no es válida"))
			return
		}
		abortionDate = &parsed
	}

	var cause *abortiondomain.Cause
	if request.Cause != nil {
		parsed, err := abortiondomain.ParseCause(*request.Cause)
		if err != nil {
			platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La causa del aborto no es válida"))
			return
		}
		cause = &parsed
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if err := h.updateAbortion.Execute(r.Context(), abortionID, abortionapplication.UpdateAbortionCommand{
		AbortionDate: abortionDate,
		Cause:        cause,
		Note:         request.Note,
		UpdatedBy:    actor.UserID,
	}); err != nil {
		writeAbortionError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// delete handles DELETE /api/v1/abortions/{id} and removes an abortion.
// It parses the id path value and maps an abortion with later sow events to 409.
func (h *AbortionHandler) delete(w http.ResponseWriter, r *http.Request) {
	abortionID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del aborto no es válido"))
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if err := h.deleteAbortion.Execute(r.Context(), abortionapplication.DeleteAbortionCommand{
		ID:        abortionID,
		DeletedBy: actor.UserID,
	}); err != nil {
		writeAbortionError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// list handles GET /api/v1/abortions and returns one page of abortions matching the filter.
// It parses the optional sow_code query parameter plus the page parameter.
func (h *AbortionHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseAbortionFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	page, err := parseAbortionPage(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	abortions, err := h.listAbortions.Execute(r.Context(), filter, page)
	if err != nil {
		writeAbortionError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toAbortionPageResponse(abortions))
}

// parseAbortionFilter reads the optional list filters from the query string.
// It returns an error for a sow code longer than 50 characters.
func parseAbortionFilter(r *http.Request) (ports.AbortionFilter, error) {
	query := r.URL.Query()
	var filter ports.AbortionFilter

	if sowCode := strings.TrimSpace(query.Get("sow_code")); sowCode != "" {
		if len([]rune(sowCode)) > maxSowCodeLength {
			return ports.AbortionFilter{}, errors.New("El código de la cerda no es válido")
		}
		filter.SowCode = &sowCode
	}

	return filter, nil
}

// parseAbortionPage reads the optional page query parameter.
// It defaults to page one and rejects non-positive or malformed values.
func parseAbortionPage(r *http.Request) (int, error) {
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

// toAbortionResponse maps a domain abortion to the HTTP response shape.
// It formats the date as YYYY-MM-DD and omits the audit fields.
func toAbortionResponse(abortion *abortiondomain.Abortion) abortionResponse {
	return abortionResponse{
		ID:           abortion.ID,
		SowID:        abortion.SowID,
		SowCode:      abortion.SowCode,
		ServiceID:    abortion.ServiceID,
		AbortionDate: abortion.AbortionDate.UTC().Format(dateLayout),
		Cause:        string(abortion.Cause),
		Note:         abortion.Note,
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

// toAbortionPageResponse maps one page of domain abortions to the paginated response shape.
// It returns an empty items slice instead of null when there are no abortions.
func toAbortionPageResponse(page abortionapplication.AbortionPage) abortionPageResponse {
	return abortionPageResponse{
		Items:      toAbortionResponses(page.Items),
		Total:      page.Total,
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalPages: page.TotalPages,
	}
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
		errors.Is(err, abortionapplication.ErrServiceNotConfirmed),
		errors.Is(err, ports.ErrAbortionNotDeletable):
		status = http.StatusConflict
	case errors.Is(err, abortiondomain.ErrInvalidID),
		errors.Is(err, abortiondomain.ErrInvalidSow),
		errors.Is(err, abortiondomain.ErrInvalidService),
		errors.Is(err, abortiondomain.ErrInvalidAbortionDate),
		errors.Is(err, abortiondomain.ErrInvalidLastMount),
		errors.Is(err, abortiondomain.ErrInvalidCause),
		errors.Is(err, abortiondomain.ErrInvalidNote),
		errors.Is(err, abortiondomain.ErrInvalidUpdate),
		errors.Is(err, abortiondomain.ErrInvalidCreatedBy),
		errors.Is(err, abortiondomain.ErrInvalidUpdatedBy),
		errors.Is(err, abortiondomain.ErrAbortionDateBeforeMount),
		errors.Is(err, abortiondomain.ErrAbortionDateBeforeEntry),
		errors.Is(err, abortiondomain.ErrAbortionDateInFuture):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
