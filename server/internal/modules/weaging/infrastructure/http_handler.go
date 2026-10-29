package infrastructure

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	farrowingdomain "server/internal/modules/farrowing/domain"
	weagingapplication "server/internal/modules/weaging/application"
	weagingdomain "server/internal/modules/weaging/domain"
	"server/internal/modules/weaging/ports"
	platformhttp "server/internal/platform/http"
)

const (
	dateLayout       = "2006-01-02"
	maxSowCodeLength = 50
)

type CreateWeagingUseCase interface {
	Execute(context.Context, weagingapplication.CreateWeagingCommand) (*weagingdomain.Weaging, error)
}

type ListWeagingsUseCase interface {
	Execute(context.Context, ports.WeagingFilter, int) (weagingapplication.WeagingPage, error)
}

type DeleteWeagingUseCase interface {
	Execute(context.Context, weagingapplication.DeleteWeagingCommand) error
}

type UpdateWeagingUseCase interface {
	Execute(context.Context, uuid.UUID, weagingapplication.UpdateWeagingCommand) error
}

type WeagingHandler struct {
	createWeaging   CreateWeagingUseCase
	listWeagings    ListWeagingsUseCase
	updateWeaging   UpdateWeagingUseCase
	deleteWeaging   DeleteWeagingUseCase
	authMiddleware  func(http.Handler) http.Handler
	adminMiddleware func(http.Handler) http.Handler
}

// NewWeagingHandler wires the weaging use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewWeagingHandler(
	createWeaging CreateWeagingUseCase,
	listWeagings ListWeagingsUseCase,
	updateWeaging UpdateWeagingUseCase,
	deleteWeaging DeleteWeagingUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *WeagingHandler {
	return &WeagingHandler{
		createWeaging:   createWeaging,
		listWeagings:    listWeagings,
		updateWeaging:   updateWeaging,
		deleteWeaging:   deleteWeaging,
		authMiddleware:  authMiddleware,
		adminMiddleware: adminMiddleware,
	}
}

// RegisterRoutes registers the weaging create, list, update and delete endpoints on the mux.
// Creation, update and deletion are admin-only while the list route only requires authentication.
func (h *WeagingHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/weagings", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/weagings", h.authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("PATCH /api/v1/weagings/{id}", h.adminMiddleware(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /api/v1/weagings/{id}", h.adminMiddleware(http.HandlerFunc(h.delete)))
}

type createWeagingRequest struct {
	SowID       string   `json:"sow_id"`
	WeagingDate string   `json:"weaging_date"`
	Quantity    int      `json:"quantity"`
	TotalWeight *float64 `json:"total_weight"`
	Destination *string  `json:"destination"`
	Note        *string  `json:"note"`
}

type updateWeagingRequest struct {
	TotalWeight      *float64 `json:"total_weight"`
	ClearTotalWeight bool     `json:"clear_total_weight"`
	Destination      *string  `json:"destination"`
	Note             *string  `json:"note"`
}

type weagingResponse struct {
	ID          uuid.UUID `json:"id"`
	FarrowingID uuid.UUID `json:"farrowing_id"`
	SowID       uuid.UUID `json:"sow_id"`
	SowCode     string    `json:"sow_code"`
	WeagingDate string    `json:"weaging_date"`
	Quantity    int       `json:"quantity"`
	TotalWeight *float64  `json:"total_weight"`
	Destination *string   `json:"destination"`
	Note        *string   `json:"note"`
}

// weagingPageResponse is one page of weagings together with its pagination metadata.
type weagingPageResponse struct {
	Items      []weagingResponse `json:"items"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
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

	if _, err := h.createWeaging.Execute(r.Context(), weagingapplication.CreateWeagingCommand{
		SowID:       sowID,
		WeagingDate: weagingDate,
		Quantity:    request.Quantity,
		TotalWeight: request.TotalWeight,
		Destination: request.Destination,
		Note:        request.Note,
		CreatedBy:   actor.UserID,
	}); err != nil {
		writeWeagingError(w, err)
		return
	}

	platformhttp.WriteStatus(w, http.StatusCreated)
}

// list handles GET /api/v1/weagings and returns one page of weagings matching the filter.
// It parses the optional sow_code, from and to query parameters plus the page parameter.
func (h *WeagingHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseWeagingFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	page, err := parseWeagingPage(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	weagings, err := h.listWeagings.Execute(r.Context(), filter, page)
	if err != nil {
		writeWeagingError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toWeagingPageResponse(weagings))
}

// parseWeagingFilter reads the optional list filters from the query string.
// It rejects a sow code longer than 50 characters, a malformed date and a range
// whose start is later than its end; sow_code narrows by the related sow code.
func parseWeagingFilter(r *http.Request) (ports.WeagingFilter, error) {
	query := r.URL.Query()
	var filter ports.WeagingFilter

	if sowCode := strings.TrimSpace(query.Get("sow_code")); sowCode != "" {
		if utf8.RuneCountInString(sowCode) > maxSowCodeLength {
			return ports.WeagingFilter{}, errors.New("El código de la cerda debe tener como máximo 50 caracteres")
		}
		filter.SowCode = &sowCode
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

// parseWeagingPage reads the optional page query parameter.
// It defaults to page one and rejects non-positive or malformed values.
func parseWeagingPage(r *http.Request) (int, error) {
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

// toWeagingResponse maps a domain weaging to the HTTP response shape.
// It formats the weaging date as YYYY-MM-DD and omits the audit fields.
func toWeagingResponse(weaging *weagingdomain.Weaging) weagingResponse {
	return weagingResponse{
		ID:          weaging.ID,
		FarrowingID: weaging.FarrowingID,
		SowID:       weaging.SowID,
		SowCode:     weaging.SowCode,
		WeagingDate: weaging.WeagingDate.UTC().Format(dateLayout),
		Quantity:    weaging.Quantity,
		TotalWeight: weaging.TotalWeight,
		Destination: weaging.Destination,
		Note:        weaging.Note,
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

// toWeagingPageResponse maps one page of domain weagings to the paginated shape.
// It returns an empty items slice instead of null when the page has no weagings.
func toWeagingPageResponse(page weagingapplication.WeagingPage) weagingPageResponse {
	return weagingPageResponse{
		Items:      toWeagingResponses(page.Items),
		Total:      page.Total,
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalPages: page.TotalPages,
	}
}

// update handles PATCH /api/v1/weagings/{id} and applies the editable fields.
// It parses the id, weight, destination and note, never accepts the sow, date or
// quantity, reads the actor from the context and responds with 200 without a body.
func (h *WeagingHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del destete no es válido"))
		return
	}

	var request updateWeagingRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if err := h.updateWeaging.Execute(r.Context(), id, weagingapplication.UpdateWeagingCommand{
		TotalWeight:      request.TotalWeight,
		ClearTotalWeight: request.ClearTotalWeight,
		Destination:      request.Destination,
		Note:             request.Note,
		UpdatedBy:        actor.UserID,
	}); err != nil {
		writeWeagingError(w, err)
		return
	}

	platformhttp.WriteStatus(w, http.StatusOK)
}

// delete handles DELETE /api/v1/weagings/{id} and removes a weaging.
// It parses the id path value, reads the actor from the context, lets the use case
// restore the farrowing and sow, and responds with no content on success.
func (h *WeagingHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del destete no es válido"))
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if err := h.deleteWeaging.Execute(r.Context(), weagingapplication.DeleteWeagingCommand{
		ID:        id,
		DeletedBy: actor.UserID,
	}); err != nil {
		writeWeagingError(w, err)
		return
	}

	platformhttp.WriteStatus(w, http.StatusNoContent)
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
		errors.Is(err, farrowingdomain.ErrWeagingQuantityMismatch),
		errors.Is(err, weagingdomain.ErrHasFutureEvents):
		status = http.StatusConflict
	case errors.Is(err, weagingdomain.ErrInvalidID),
		errors.Is(err, weagingdomain.ErrInvalidFarrowing),
		errors.Is(err, weagingdomain.ErrInvalidSow),
		errors.Is(err, weagingdomain.ErrInvalidWeagingDate),
		errors.Is(err, weagingdomain.ErrInvalidFarrowDate),
		errors.Is(err, weagingdomain.ErrInvalidQuantity),
		errors.Is(err, weagingdomain.ErrQuantityExceedsMax),
		errors.Is(err, weagingdomain.ErrInvalidTotalWeight),
		errors.Is(err, weagingdomain.ErrTotalWeightExceedsMax),
		errors.Is(err, weagingdomain.ErrInvalidDestination),
		errors.Is(err, weagingdomain.ErrInvalidNote),
		errors.Is(err, weagingdomain.ErrInvalidCreatedBy),
		errors.Is(err, weagingdomain.ErrInvalidUpdatedBy),
		errors.Is(err, weagingdomain.ErrWeagingDateBeforeFarrowing),
		errors.Is(err, weagingdomain.ErrWeagingDateBeforeEvents),
		errors.Is(err, weagingdomain.ErrWeagingDateInFuture),
		errors.Is(err, weagingdomain.ErrInvalidUpdate),
		errors.Is(err, farrowingdomain.ErrInvalidPigletQuantity):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
