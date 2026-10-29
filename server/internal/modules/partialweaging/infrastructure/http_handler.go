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
	farrowingdomain "server/internal/modules/farrowing/domain"
	partialweagingapplication "server/internal/modules/partialweaging/application"
	partialweagingdomain "server/internal/modules/partialweaging/domain"
	"server/internal/modules/partialweaging/ports"
	platformhttp "server/internal/platform/http"
)

const (
	dateLayout = "2006-01-02"
	// maxSowCodeLength mirrors the length of the sows.code column.
	maxSowCodeLength = 50
)

type CreatePartialWeagingUseCase interface {
	Execute(context.Context, partialweagingapplication.CreatePartialWeagingCommand) (*partialweagingdomain.PartialWeaging, error)
}

type ListPartialWeagingsUseCase interface {
	Execute(context.Context, ports.PartialWeagingFilter, int) (partialweagingapplication.PartialWeagingPage, error)
}

type DeletePartialWeagingUseCase interface {
	Execute(context.Context, partialweagingapplication.DeletePartialWeagingCommand) error
}

type UpdatePartialWeagingUseCase interface {
	Execute(context.Context, uuid.UUID, partialweagingapplication.UpdatePartialWeagingCommand) error
}

type PartialWeagingHandler struct {
	createPartialWeaging CreatePartialWeagingUseCase
	listPartialWeagings  ListPartialWeagingsUseCase
	deletePartialWeaging DeletePartialWeagingUseCase
	updatePartialWeaging UpdatePartialWeagingUseCase
	authMiddleware       func(http.Handler) http.Handler
	adminMiddleware      func(http.Handler) http.Handler
}

// NewPartialWeagingHandler wires the partial weaging use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewPartialWeagingHandler(
	createPartialWeaging CreatePartialWeagingUseCase,
	listPartialWeagings ListPartialWeagingsUseCase,
	deletePartialWeaging DeletePartialWeagingUseCase,
	updatePartialWeaging UpdatePartialWeagingUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *PartialWeagingHandler {
	return &PartialWeagingHandler{
		createPartialWeaging: createPartialWeaging,
		listPartialWeagings:  listPartialWeagings,
		deletePartialWeaging: deletePartialWeaging,
		updatePartialWeaging: updatePartialWeaging,
		authMiddleware:       authMiddleware,
		adminMiddleware:      adminMiddleware,
	}
}

// RegisterRoutes registers the partial weaging create, list, update and delete endpoints on the mux.
// Creation, update and deletion are admin-only while the list route only requires authentication.
func (h *PartialWeagingHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/partial-weagings", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/partial-weagings", h.authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("PATCH /api/v1/partial-weagings/{id}", h.adminMiddleware(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /api/v1/partial-weagings/{id}", h.adminMiddleware(http.HandlerFunc(h.delete)))
}

type createPartialWeagingRequest struct {
	SowID       string   `json:"sow_id"`
	WeagingDate string   `json:"weaging_date"`
	Quantity    int      `json:"quantity"`
	TotalWeight *float64 `json:"total_weight"`
	Type        string   `json:"type"`
	Note        *string  `json:"note"`
}

type updatePartialWeagingRequest struct {
	TotalWeight      *float64 `json:"total_weight"`
	ClearTotalWeight bool     `json:"clear_total_weight"`
	Note             *string  `json:"note"`
}

type partialWeagingResponse struct {
	ID          uuid.UUID `json:"id"`
	FarrowingID uuid.UUID `json:"farrowing_id"`
	SowID       uuid.UUID `json:"sow_id"`
	SowCode     string    `json:"sow_code"`
	WeagingDate string    `json:"weaging_date"`
	Quantity    int       `json:"quantity"`
	TotalWeight *float64  `json:"total_weight"`
	Type        string    `json:"type"`
	Note        *string   `json:"note"`
}

// partialWeagingPageResponse is one page of partial weagings with its pagination metadata.
type partialWeagingPageResponse struct {
	Items      []partialWeagingResponse `json:"items"`
	Total      int                      `json:"total"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	TotalPages int                      `json:"total_pages"`
}

// create handles POST /api/v1/partial-weagings and registers a partial weaging.
// It parses the sow, weaging date, quantity, weight, type and note, reads the actor
// from the context, lets the use case derive the latest farrowing and responds with
// the created status only, without a body.
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

	if _, err := h.createPartialWeaging.Execute(r.Context(), partialweagingapplication.CreatePartialWeagingCommand{
		SowID:       sowID,
		WeagingDate: weagingDate,
		Quantity:    request.Quantity,
		TotalWeight: request.TotalWeight,
		Type:        weagingType,
		Note:        request.Note,
		CreatedBy:   actor.UserID,
	}); err != nil {
		writePartialWeagingError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// update handles PATCH /api/v1/partial-weagings/{id} and applies the editable fields.
// It parses the id, weight and note, never accepts the sow, date, quantity or type,
// reads the actor from the context and responds with 200 without a body.
func (h *PartialWeagingHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del destete parcial no es válido"))
		return
	}

	var request updatePartialWeagingRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if err := h.updatePartialWeaging.Execute(r.Context(), id, partialweagingapplication.UpdatePartialWeagingCommand{
		TotalWeight:      request.TotalWeight,
		ClearTotalWeight: request.ClearTotalWeight,
		Note:             request.Note,
		UpdatedBy:        actor.UserID,
	}); err != nil {
		writePartialWeagingError(w, err)
		return
	}

	platformhttp.WriteStatus(w, http.StatusOK)
}

// delete handles DELETE /api/v1/partial-weagings/{id} and removes a partial weaging.
// It parses the id path value, reads the actor from the context, lets the use case
// restore the farrowing and sow, and responds with no content on success.
func (h *PartialWeagingHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del destete parcial no es válido"))
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if err := h.deletePartialWeaging.Execute(r.Context(), partialweagingapplication.DeletePartialWeagingCommand{
		ID:        id,
		DeletedBy: actor.UserID,
	}); err != nil {
		writePartialWeagingError(w, err)
		return
	}

	platformhttp.WriteStatus(w, http.StatusNoContent)
}

// list handles GET /api/v1/partial-weagings and returns one page matching the filter.
// It parses the optional sow_code, from and to query parameters plus the page parameter.
func (h *PartialWeagingHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parsePartialWeagingFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	page, err := parsePartialWeagingPage(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	weagings, err := h.listPartialWeagings.Execute(r.Context(), filter, page)
	if err != nil {
		writePartialWeagingError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toPartialWeagingPageResponse(weagings))
}

// parsePartialWeagingFilter reads the optional list filters from the query string.
// It returns an error for a sow code over 50 characters, a malformed date value or
// a date range whose start is later than its end.
func parsePartialWeagingFilter(r *http.Request) (ports.PartialWeagingFilter, error) {
	query := r.URL.Query()
	var filter ports.PartialWeagingFilter

	if sowCode := strings.TrimSpace(query.Get("sow_code")); sowCode != "" {
		if len([]rune(sowCode)) > maxSowCodeLength {
			return ports.PartialWeagingFilter{}, errors.New("El código de la cerda no puede superar los 50 caracteres")
		}
		filter.SowCode = &sowCode
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

// parsePartialWeagingPage reads the optional page query parameter.
// It defaults to page one and rejects non-positive or malformed values.
func parsePartialWeagingPage(r *http.Request) (int, error) {
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

// toPartialWeagingResponse maps a domain partial weaging to the HTTP response shape.
// It formats the weaging date as YYYY-MM-DD, exposes the sow code and omits audit fields.
func toPartialWeagingResponse(weaging *partialweagingdomain.PartialWeaging) partialWeagingResponse {
	return partialWeagingResponse{
		ID:          weaging.ID,
		FarrowingID: weaging.FarrowingID,
		SowID:       weaging.SowID,
		SowCode:     weaging.SowCode,
		WeagingDate: weaging.WeagingDate.UTC().Format(dateLayout),
		Quantity:    weaging.Quantity,
		TotalWeight: weaging.TotalWeight,
		Type:        string(weaging.Type),
		Note:        weaging.Note,
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

// toPartialWeagingPageResponse maps one page of domain partial weagings to the paginated shape.
// It returns an empty items slice instead of null when the page has no partial weagings.
func toPartialWeagingPageResponse(page partialweagingapplication.PartialWeagingPage) partialWeagingPageResponse {
	return partialWeagingPageResponse{
		Items:      toPartialWeagingResponses(page.Items),
		Total:      page.Total,
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalPages: page.TotalPages,
	}
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
		errors.Is(err, farrowingdomain.ErrInsufficientPiglets),
		errors.Is(err, partialweagingdomain.ErrHasFutureEvents),
		errors.Is(err, partialweagingdomain.ErrHasLaterNurseEvents):
		status = http.StatusConflict
	case errors.Is(err, partialweagingdomain.ErrInvalidID),
		errors.Is(err, partialweagingdomain.ErrInvalidFarrowing),
		errors.Is(err, partialweagingdomain.ErrInvalidSow),
		errors.Is(err, partialweagingdomain.ErrInvalidWeagingDate),
		errors.Is(err, partialweagingdomain.ErrInvalidFarrowDate),
		errors.Is(err, partialweagingdomain.ErrInvalidQuantity),
		errors.Is(err, partialweagingdomain.ErrQuantityExceedsMax),
		errors.Is(err, partialweagingdomain.ErrQuantityMismatch),
		errors.Is(err, partialweagingdomain.ErrInvalidTotalWeight),
		errors.Is(err, partialweagingdomain.ErrTotalWeightExceedsMax),
		errors.Is(err, partialweagingdomain.ErrInvalidType),
		errors.Is(err, partialweagingdomain.ErrInvalidNote),
		errors.Is(err, partialweagingdomain.ErrInvalidCreatedBy),
		errors.Is(err, partialweagingdomain.ErrInvalidUpdatedBy),
		errors.Is(err, partialweagingdomain.ErrInvalidUpdate),
		errors.Is(err, partialweagingdomain.ErrWeagingDateBeforeFarrowing),
		errors.Is(err, partialweagingdomain.ErrWeagingDateBeforeEvents),
		errors.Is(err, partialweagingdomain.ErrWeagingDateBeforePartial),
		errors.Is(err, partialweagingdomain.ErrWeagingDateInFuture),
		errors.Is(err, farrowingdomain.ErrInvalidPigletQuantity):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
