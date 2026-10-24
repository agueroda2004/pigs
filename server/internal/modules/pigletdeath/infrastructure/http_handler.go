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
	pigletdeathapplication "server/internal/modules/pigletdeath/application"
	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	"server/internal/modules/pigletdeath/ports"
	platformhttp "server/internal/platform/http"
)

const (
	dateLayout = "2006-01-02"
	// maxSowCodeLength mirrors the length of the sows.code column.
	maxSowCodeLength = 50
)

type CreatePigletDeathUseCase interface {
	Execute(context.Context, pigletdeathapplication.CreatePigletDeathCommand) (*pigletdeathdomain.PigletDeath, error)
}

type ListPigletDeathsUseCase interface {
	Execute(context.Context, ports.PigletDeathFilter, int) (pigletdeathapplication.PigletDeathPage, error)
}

type DeletePigletDeathUseCase interface {
	Execute(context.Context, pigletdeathapplication.DeletePigletDeathCommand) error
}

type UpdatePigletDeathUseCase interface {
	Execute(context.Context, uuid.UUID, pigletdeathapplication.UpdatePigletDeathCommand) (*pigletdeathdomain.PigletDeath, error)
}

type PigletDeathHandler struct {
	createPigletDeath CreatePigletDeathUseCase
	listPigletDeaths  ListPigletDeathsUseCase
	deletePigletDeath DeletePigletDeathUseCase
	updatePigletDeath UpdatePigletDeathUseCase
	authMiddleware    func(http.Handler) http.Handler
	adminMiddleware   func(http.Handler) http.Handler
}

// NewPigletDeathHandler wires the piglet death use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewPigletDeathHandler(
	createPigletDeath CreatePigletDeathUseCase,
	listPigletDeaths ListPigletDeathsUseCase,
	deletePigletDeath DeletePigletDeathUseCase,
	updatePigletDeath UpdatePigletDeathUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *PigletDeathHandler {
	return &PigletDeathHandler{
		createPigletDeath: createPigletDeath,
		listPigletDeaths:  listPigletDeaths,
		deletePigletDeath: deletePigletDeath,
		updatePigletDeath: updatePigletDeath,
		authMiddleware:    authMiddleware,
		adminMiddleware:   adminMiddleware,
	}
}

// RegisterRoutes registers the piglet death create, list, update and delete endpoints.
// Create, update and delete are admin-only while the list route only requires authentication.
func (h *PigletDeathHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/piglet-deaths", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/piglet-deaths", h.authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("PATCH /api/v1/piglet-deaths/{id}", h.adminMiddleware(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /api/v1/piglet-deaths/{id}", h.adminMiddleware(http.HandlerFunc(h.delete)))
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

type updatePigletDeathRequest struct {
	DeathDate  *string  `json:"death_date"`
	Quantity   *int     `json:"quantity"`
	SowID      *string  `json:"sow_id"`
	OperatorID *string  `json:"operator_id"`
	Weight     *float64 `json:"weight"`
	Cause      *string  `json:"cause"`
	Turn       *string  `json:"turn"`
	Note       *string  `json:"note"`
}

type pigletDeathResponse struct {
	ID           uuid.UUID `json:"id"`
	FarrowingID  uuid.UUID `json:"farrowing_id"`
	SowID        uuid.UUID `json:"sow_id"`
	SowCode      string    `json:"sow_code"`
	OperatorID   uuid.UUID `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	DeathDate    string    `json:"death_date"`
	Quantity     int       `json:"quantity"`
	Weight       *float64  `json:"weight"`
	Cause        string    `json:"cause"`
	Turn         string    `json:"turn"`
	Note         *string   `json:"note"`
}

// pigletDeathPageResponse is one page of piglet deaths with its pagination metadata.
type pigletDeathPageResponse struct {
	Items      []pigletDeathResponse `json:"items"`
	Total      int                   `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

// create handles POST /api/v1/piglet-deaths and registers a piglet death.
// It parses the sow, operator, date, quantity, cause and turn, reads the actor
// from the context, lets the use case resolve the sow's latest farrowing and
// responds with 201 and no body.
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

	_, err = h.createPigletDeath.Execute(r.Context(), pigletdeathapplication.CreatePigletDeathCommand{
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

	platformhttp.WriteStatus(w, http.StatusCreated)
}

// list handles GET /api/v1/piglet-deaths and returns one page of deaths matching the filter.
// It parses the optional sow_code, from and to query parameters plus the page parameter.
func (h *PigletDeathHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parsePigletDeathFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	page, err := parsePigletDeathPage(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	deaths, err := h.listPigletDeaths.Execute(r.Context(), filter, page)
	if err != nil {
		writePigletDeathError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toPigletDeathPageResponse(deaths))
}

// update handles PATCH /api/v1/piglet-deaths/{id} and applies the editable fields.
// It parses the id, reads the actor to set UpdatedBy and answers with a status code only.
func (h *PigletDeathHandler) update(w http.ResponseWriter, r *http.Request) {
	deathID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La muerte de lechones no es válida"))
		return
	}

	var request updatePigletDeathRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	params, err := parseUpdatePigletDeathParams(request)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if _, err := h.updatePigletDeath.Execute(r.Context(), deathID, pigletdeathapplication.UpdatePigletDeathCommand{
		Params:    params,
		UpdatedBy: actor.UserID,
	}); err != nil {
		writePigletDeathError(w, err)
		return
	}

	platformhttp.WriteStatus(w, http.StatusOK)
}

// parseUpdatePigletDeathParams converts the raw update request into domain params.
// It parses the optional date, sow, operator, cause and turn and rejects malformed values.
func parseUpdatePigletDeathParams(request updatePigletDeathRequest) (pigletdeathdomain.UpdatePigletDeathParams, error) {
	var params pigletdeathdomain.UpdatePigletDeathParams
	params.Quantity = request.Quantity
	params.Weight = request.Weight
	params.Note = request.Note

	if request.DeathDate != nil {
		deathDate, err := time.Parse(dateLayout, *request.DeathDate)
		if err != nil {
			return params, errors.New("La fecha de la muerte no es válida")
		}
		params.DeathDate = &deathDate
	}

	if request.SowID != nil {
		sowID, err := uuid.Parse(*request.SowID)
		if err != nil {
			return params, errors.New("La cerda no es válida")
		}
		params.SowID = &sowID
	}

	if request.OperatorID != nil {
		operatorID, err := uuid.Parse(*request.OperatorID)
		if err != nil {
			return params, errors.New("El operador no es válido")
		}
		params.OperatorID = &operatorID
	}

	if request.Cause != nil {
		cause, err := pigletdeathdomain.ParseCause(*request.Cause)
		if err != nil {
			return params, errors.New("La causa de la muerte no es válida")
		}
		params.Cause = &cause
	}

	if request.Turn != nil {
		turn, err := pigletdeathdomain.ParseTurn(*request.Turn)
		if err != nil {
			return params, errors.New("El turno no es válido")
		}
		params.Turn = &turn
	}

	return params, nil
}

// delete handles DELETE /api/v1/piglet-deaths/{id} and removes a piglet death.
// It parses the id and the actor from the context and responds with 204 and no body.
func (h *PigletDeathHandler) delete(w http.ResponseWriter, r *http.Request) {
	deathID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La muerte de lechones no es válida"))
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if err := h.deletePigletDeath.Execute(r.Context(), pigletdeathapplication.DeletePigletDeathCommand{
		ID:        deathID,
		DeletedBy: actor.UserID,
	}); err != nil {
		writePigletDeathError(w, err)
		return
	}

	platformhttp.WriteStatus(w, http.StatusNoContent)
}

// parsePigletDeathFilter reads the optional list filters from the query string.
// It returns an error for a malformed sow_code, from or to value and rejects a
// sow code over 50 characters or a date range whose start is later than its end.
func parsePigletDeathFilter(r *http.Request) (ports.PigletDeathFilter, error) {
	query := r.URL.Query()
	var filter ports.PigletDeathFilter

	if sowCode := strings.TrimSpace(query.Get("sow_code")); sowCode != "" {
		if len([]rune(sowCode)) > maxSowCodeLength {
			return ports.PigletDeathFilter{}, errors.New("El código de la cerda no puede superar los 50 caracteres")
		}
		filter.SowCode = &sowCode
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

// parsePigletDeathPage reads the optional page query parameter.
// It defaults to page one and rejects non-positive or malformed values.
func parsePigletDeathPage(r *http.Request) (int, error) {
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

// toPigletDeathResponse maps a domain piglet death to the HTTP response shape.
// It formats the date as YYYY-MM-DD and exposes the related sow code.
func toPigletDeathResponse(death *pigletdeathdomain.PigletDeath) pigletDeathResponse {
	return pigletDeathResponse{
		ID:           death.ID,
		FarrowingID:  death.FarrowingID,
		SowID:        death.SowID,
		SowCode:      death.SowCode,
		OperatorID:   death.OperatorID,
		OperatorName: death.OperatorName,
		DeathDate:    death.DeathDate.UTC().Format(dateLayout),
		Quantity:     death.Quantity,
		Weight:       death.Weight,
		Cause:        string(death.Cause),
		Turn:         string(death.Turn),
		Note:         death.Note,
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

// toPigletDeathPageResponse maps one page of domain piglet deaths to the paginated shape.
// It returns an empty items slice instead of null when the page has no deaths.
func toPigletDeathPageResponse(page pigletdeathapplication.PigletDeathPage) pigletDeathPageResponse {
	return pigletDeathPageResponse{
		Items:      toPigletDeathResponses(page.Items),
		Total:      page.Total,
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalPages: page.TotalPages,
	}
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
		errors.Is(err, farrowingdomain.ErrInsufficientPiglets),
		errors.Is(err, pigletdeathdomain.ErrHasLaterWeaging),
		errors.Is(err, pigletdeathdomain.ErrHasLaterPartialWeaging):
		status = http.StatusConflict
	case errors.Is(err, pigletdeathdomain.ErrInvalidID),
		errors.Is(err, pigletdeathdomain.ErrInvalidFarrowing),
		errors.Is(err, pigletdeathdomain.ErrInvalidSow),
		errors.Is(err, pigletdeathdomain.ErrInvalidOperator),
		errors.Is(err, pigletdeathdomain.ErrInvalidDeathDate),
		errors.Is(err, pigletdeathdomain.ErrInvalidFarrowDate),
		errors.Is(err, pigletdeathdomain.ErrInvalidQuantity),
		errors.Is(err, pigletdeathdomain.ErrQuantityTooHigh),
		errors.Is(err, pigletdeathdomain.ErrWeightRequired),
		errors.Is(err, pigletdeathdomain.ErrInvalidWeight),
		errors.Is(err, pigletdeathdomain.ErrWeightTooHigh),
		errors.Is(err, pigletdeathdomain.ErrInvalidCause),
		errors.Is(err, pigletdeathdomain.ErrInvalidTurn),
		errors.Is(err, pigletdeathdomain.ErrInvalidNote),
		errors.Is(err, pigletdeathdomain.ErrInvalidCreatedBy),
		errors.Is(err, pigletdeathdomain.ErrInvalidUpdatedBy),
		errors.Is(err, pigletdeathdomain.ErrDeathDateBeforeFarrowing),
		errors.Is(err, pigletdeathdomain.ErrDeathDateInFuture),
		errors.Is(err, pigletdeathdomain.ErrDeathDateBeforeWeaging),
		errors.Is(err, pigletdeathdomain.ErrInvalidUpdate),
		errors.Is(err, pigletdeathdomain.ErrImmutableDeathField),
		errors.Is(err, farrowingdomain.ErrInvalidPigletQuantity):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
