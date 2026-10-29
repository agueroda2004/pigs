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
	pigletfosteringapplication "server/internal/modules/pigletfostering/application"
	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
	"server/internal/modules/pigletfostering/ports"
	platformhttp "server/internal/platform/http"
)

const (
	dateLayout = "2006-01-02"
	// maxSowCodeLength mirrors the length of the sows.code column.
	maxSowCodeLength = 50
)

type CreatePigletFosteringUseCase interface {
	Execute(context.Context, pigletfosteringapplication.CreatePigletFosteringCommand) (*pigletfosteringdomain.PigletFostering, error)
}

type UpdatePigletFosteringUseCase interface {
	Execute(context.Context, uuid.UUID, pigletfosteringapplication.UpdatePigletFosteringCommand) (*pigletfosteringdomain.PigletFostering, error)
}

type DeletePigletFosteringUseCase interface {
	Execute(context.Context, uuid.UUID, pigletfosteringapplication.DeletePigletFosteringCommand) error
}

type ListPigletFosteringsUseCase interface {
	Execute(context.Context, ports.PigletFosteringFilter, int) (pigletfosteringapplication.PigletFosteringPage, error)
}

type PigletFosteringHandler struct {
	createPigletFostering CreatePigletFosteringUseCase
	updatePigletFostering UpdatePigletFosteringUseCase
	deletePigletFostering DeletePigletFosteringUseCase
	listPigletFosterings  ListPigletFosteringsUseCase
	authMiddleware        func(http.Handler) http.Handler
	adminMiddleware       func(http.Handler) http.Handler
}

// NewPigletFosteringHandler wires the fostering use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewPigletFosteringHandler(
	createPigletFostering CreatePigletFosteringUseCase,
	updatePigletFostering UpdatePigletFosteringUseCase,
	deletePigletFostering DeletePigletFosteringUseCase,
	listPigletFosterings ListPigletFosteringsUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *PigletFosteringHandler {
	return &PigletFosteringHandler{
		createPigletFostering: createPigletFostering,
		updatePigletFostering: updatePigletFostering,
		deletePigletFostering: deletePigletFostering,
		listPigletFosterings:  listPigletFosterings,
		authMiddleware:        authMiddleware,
		adminMiddleware:       adminMiddleware,
	}
}

// RegisterRoutes registers the fostering create, update, delete and list endpoints.
// Creation, update and delete are admin-only while the list route only requires
// authentication.
func (h *PigletFosteringHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/piglet-fosterings", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("PATCH /api/v1/piglet-fosterings/{id}", h.adminMiddleware(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /api/v1/piglet-fosterings/{id}", h.adminMiddleware(http.HandlerFunc(h.delete)))
	mux.Handle("GET /api/v1/piglet-fosterings", h.authMiddleware(http.HandlerFunc(h.list)))
}

type createPigletFosteringRequest struct {
	DonorSowID    string  `json:"donor_sow_id"`
	ReceiverSowID string  `json:"receiver_sow_id"`
	MovementDate  string  `json:"movement_date"`
	Quantity      int     `json:"quantity"`
	Note          *string `json:"note"`
}

type updatePigletFosteringRequest struct {
	Note *string `json:"note"`
}

type pigletFosteringResponse struct {
	ID                  uuid.UUID `json:"id"`
	DonorFarrowingID    uuid.UUID `json:"donor_farrowing_id"`
	ReceiverFarrowingID uuid.UUID `json:"receiver_farrowing_id"`
	DonorSowID          uuid.UUID `json:"donor_sow_id"`
	DonorSowCode        string    `json:"donor_sow_code"`
	ReceiverSowID       uuid.UUID `json:"receiver_sow_id"`
	ReceiverSowCode     string    `json:"receiver_sow_code"`
	MovementDate        string    `json:"movement_date"`
	Quantity            int       `json:"quantity"`
	Note                *string   `json:"note"`
}

// pigletFosteringPageResponse is one page of fosterings with its pagination metadata.
type pigletFosteringPageResponse struct {
	Items      []pigletFosteringResponse `json:"items"`
	Total      int                       `json:"total"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
	TotalPages int                       `json:"total_pages"`
}

// create handles POST /api/v1/piglet-fosterings and registers a fostering.
// It parses both sows, the movement date, quantity and note, reads the actor from
// the context and lets the use case resolve each sow's latest farrowing.
func (h *PigletFosteringHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createPigletFosteringRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	donorSowID, err := uuid.Parse(request.DonorSowID)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La cerda donante no es válida"))
		return
	}

	receiverSowID, err := uuid.Parse(request.ReceiverSowID)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La cerda receptora no es válida"))
		return
	}

	movementDate, err := time.Parse(dateLayout, request.MovementDate)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La fecha del traslado no es válida"))
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	_, err = h.createPigletFostering.Execute(r.Context(), pigletfosteringapplication.CreatePigletFosteringCommand{
		DonorSowID:    donorSowID,
		ReceiverSowID: receiverSowID,
		MovementDate:  movementDate,
		Quantity:      request.Quantity,
		Note:          request.Note,
		CreatedBy:     actor.UserID,
	})
	if err != nil {
		writePigletFosteringError(w, err)
		return
	}

	platformhttp.WriteStatus(w, http.StatusCreated)
}

// update handles PATCH /api/v1/piglet-fosterings/{id} and applies the editable fields.
// It parses the id and the optional note, reads the actor and answers with a status
// code only. Date, quantity, sow and farrowing changes are rejected.
func (h *PigletFosteringHandler) update(w http.ResponseWriter, r *http.Request) {
	fosteringID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El traslado de lechones no es válido"))
		return
	}

	var request updatePigletFosteringRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	params := parseUpdatePigletFosteringParams(request)

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if _, err := h.updatePigletFostering.Execute(r.Context(), fosteringID, pigletfosteringapplication.UpdatePigletFosteringCommand{
		Params:    params,
		UpdatedBy: actor.UserID,
	}); err != nil {
		writePigletFosteringError(w, err)
		return
	}

	platformhttp.WriteStatus(w, http.StatusOK)
}

// delete handles DELETE /api/v1/piglet-fosterings/{id} and removes a fostering.
// It parses the id and the actor from the context and responds with 204 and no body.
func (h *PigletFosteringHandler) delete(w http.ResponseWriter, r *http.Request) {
	fosteringID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El traslado de lechones no es válido"))
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if err := h.deletePigletFostering.Execute(r.Context(), fosteringID, pigletfosteringapplication.DeletePigletFosteringCommand{
		DeletedBy: actor.UserID,
	}); err != nil {
		writePigletFosteringError(w, err)
		return
	}

	platformhttp.WriteStatus(w, http.StatusNoContent)
}

// parseUpdatePigletFosteringParams converts the raw update request into domain params.
// Only the note is accepted; any other field is rejected by the JSON decoder.
func parseUpdatePigletFosteringParams(request updatePigletFosteringRequest) pigletfosteringdomain.UpdatePigletFosteringParams {
	return pigletfosteringdomain.UpdatePigletFosteringParams{Note: request.Note}
}

// list handles GET /api/v1/piglet-fosterings and returns one page of fosterings
// matching the filter. It parses the optional sow_code, from and to query
// parameters plus the page parameter.
func (h *PigletFosteringHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parsePigletFosteringFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	page, err := parsePigletFosteringPage(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	fosterings, err := h.listPigletFosterings.Execute(r.Context(), filter, page)
	if err != nil {
		writePigletFosteringError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toPigletFosteringPageResponse(fosterings))
}

// parsePigletFosteringFilter reads the optional list filters from the query string.
// It returns an error for a sow code over 50 characters or a malformed date, and
// rejects a date range whose start is later than its end.
func parsePigletFosteringFilter(r *http.Request) (ports.PigletFosteringFilter, error) {
	query := r.URL.Query()
	var filter ports.PigletFosteringFilter

	if sowCode := strings.TrimSpace(query.Get("sow_code")); sowCode != "" {
		if len([]rune(sowCode)) > maxSowCodeLength {
			return ports.PigletFosteringFilter{}, errors.New("El código de la cerda no puede superar los 50 caracteres")
		}
		filter.SowCode = &sowCode
	}

	if from := strings.TrimSpace(query.Get("from")); from != "" {
		fromDate, err := time.Parse(dateLayout, from)
		if err != nil {
			return ports.PigletFosteringFilter{}, errors.New("La fecha inicial no es válida")
		}
		filter.FromDate = &fromDate
	}

	if to := strings.TrimSpace(query.Get("to")); to != "" {
		toDate, err := time.Parse(dateLayout, to)
		if err != nil {
			return ports.PigletFosteringFilter{}, errors.New("La fecha final no es válida")
		}
		filter.ToDate = &toDate
	}

	if filter.FromDate != nil && filter.ToDate != nil && filter.FromDate.After(*filter.ToDate) {
		return ports.PigletFosteringFilter{}, errors.New("La fecha inicial no puede ser posterior a la fecha final")
	}

	return filter, nil
}

// parsePigletFosteringPage reads the optional page query parameter.
// It defaults to page one and rejects non-positive or malformed values.
func parsePigletFosteringPage(r *http.Request) (int, error) {
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

// toPigletFosteringResponse maps a domain fostering to the HTTP response shape.
// It formats the movement date as YYYY-MM-DD, exposes both sow codes and omits
// the audit fields.
func toPigletFosteringResponse(fostering *pigletfosteringdomain.PigletFostering) pigletFosteringResponse {
	return pigletFosteringResponse{
		ID:                  fostering.ID,
		DonorFarrowingID:    fostering.DonorFarrowingID,
		ReceiverFarrowingID: fostering.ReceiverFarrowingID,
		DonorSowID:          fostering.DonorSowID,
		DonorSowCode:        fostering.DonorSowCode,
		ReceiverSowID:       fostering.ReceiverSowID,
		ReceiverSowCode:     fostering.ReceiverSowCode,
		MovementDate:        fostering.MovementDate.UTC().Format(dateLayout),
		Quantity:            fostering.Quantity,
		Note:                fostering.Note,
	}
}

// toPigletFosteringResponses maps a list of domain fosterings to HTTP response shapes.
// It returns an empty slice instead of null when there are no fosterings.
func toPigletFosteringResponses(fosterings []*pigletfosteringdomain.PigletFostering) []pigletFosteringResponse {
	responses := make([]pigletFosteringResponse, 0, len(fosterings))
	for _, fostering := range fosterings {
		responses = append(responses, toPigletFosteringResponse(fostering))
	}
	return responses
}

// toPigletFosteringPageResponse maps one page of domain fosterings to the paginated shape.
// It returns an empty items slice instead of null when the page has no fosterings.
func toPigletFosteringPageResponse(page pigletfosteringapplication.PigletFosteringPage) pigletFosteringPageResponse {
	return pigletFosteringPageResponse{
		Items:      toPigletFosteringResponses(page.Items),
		Total:      page.Total,
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalPages: page.TotalPages,
	}
}

// writePigletFosteringError maps fostering domain, application and port errors to status codes.
// It falls back to 500 for unrecognized errors.
func writePigletFosteringError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrPigletFosteringNotFound),
		errors.Is(err, ports.ErrSowNotFound),
		errors.Is(err, ports.ErrFarrowingNotFound):
		status = http.StatusNotFound
	case errors.Is(err, pigletfosteringapplication.ErrSowNotLactating),
		errors.Is(err, farrowingdomain.ErrInsufficientPiglets),
		errors.Is(err, pigletfosteringdomain.ErrHasLaterWeaging),
		errors.Is(err, pigletfosteringdomain.ErrHasLaterPartialWeaging):
		status = http.StatusConflict
	case errors.Is(err, pigletfosteringdomain.ErrInvalidID),
		errors.Is(err, pigletfosteringdomain.ErrInvalidDonor),
		errors.Is(err, pigletfosteringdomain.ErrInvalidReceiver),
		errors.Is(err, pigletfosteringdomain.ErrInvalidDonorSow),
		errors.Is(err, pigletfosteringdomain.ErrInvalidReceiverSow),
		errors.Is(err, pigletfosteringdomain.ErrInvalidMovementDate),
		errors.Is(err, pigletfosteringdomain.ErrInvalidFarrowDate),
		errors.Is(err, pigletfosteringdomain.ErrInvalidQuantity),
		errors.Is(err, pigletfosteringdomain.ErrQuantityExceedsMax),
		errors.Is(err, pigletfosteringdomain.ErrInvalidNote),
		errors.Is(err, pigletfosteringdomain.ErrInvalidCreatedBy),
		errors.Is(err, pigletfosteringdomain.ErrInvalidUpdatedBy),
		errors.Is(err, pigletfosteringdomain.ErrInvalidDeletedBy),
		errors.Is(err, pigletfosteringdomain.ErrInvalidUpdate),
		errors.Is(err, pigletfosteringdomain.ErrSameFarrowing),
		errors.Is(err, pigletfosteringdomain.ErrMovementDateBeforeFarrowing),
		errors.Is(err, pigletfosteringdomain.ErrMovementDateBeforePartialWeaging),
		errors.Is(err, pigletfosteringdomain.ErrMovementDateInFuture),
		errors.Is(err, farrowingdomain.ErrInvalidPigletQuantity):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
