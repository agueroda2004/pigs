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
	pigletfosteringapplication "server/internal/modules/pigletfostering/application"
	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
	"server/internal/modules/pigletfostering/ports"
	platformhttp "server/internal/platform/http"
)

const dateLayout = "2006-01-02"

type CreatePigletFosteringUseCase interface {
	Execute(context.Context, pigletfosteringapplication.CreatePigletFosteringCommand) (*pigletfosteringdomain.PigletFostering, error)
}

type ListPigletFosteringsUseCase interface {
	Execute(context.Context, ports.PigletFosteringFilter) ([]*pigletfosteringdomain.PigletFostering, error)
}

type PigletFosteringHandler struct {
	createPigletFostering CreatePigletFosteringUseCase
	listPigletFosterings  ListPigletFosteringsUseCase
	authMiddleware        func(http.Handler) http.Handler
	adminMiddleware       func(http.Handler) http.Handler
}

// NewPigletFosteringHandler wires the fostering use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewPigletFosteringHandler(
	createPigletFostering CreatePigletFosteringUseCase,
	listPigletFosterings ListPigletFosteringsUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *PigletFosteringHandler {
	return &PigletFosteringHandler{
		createPigletFostering: createPigletFostering,
		listPigletFosterings:  listPigletFosterings,
		authMiddleware:        authMiddleware,
		adminMiddleware:       adminMiddleware,
	}
}

// RegisterRoutes registers the fostering create and list endpoints on the mux.
// Creation is admin-only while the list route only requires authentication.
func (h *PigletFosteringHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/piglet-fosterings", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/piglet-fosterings", h.authMiddleware(http.HandlerFunc(h.list)))
}

type createPigletFosteringRequest struct {
	DonorSowID    string  `json:"donor_sow_id"`
	ReceiverSowID string  `json:"receiver_sow_id"`
	MovementDate  string  `json:"movement_date"`
	Quantity      int     `json:"quantity"`
	Note          *string `json:"note"`
}

type pigletFosteringResponse struct {
	ID                  uuid.UUID `json:"id"`
	DonorFarrowingID    uuid.UUID `json:"donor_farrowing_id"`
	ReceiverFarrowingID uuid.UUID `json:"receiver_farrowing_id"`
	DonorSowID          uuid.UUID `json:"donor_sow_id"`
	ReceiverSowID       uuid.UUID `json:"receiver_sow_id"`
	MovementDate        string    `json:"movement_date"`
	Quantity            int       `json:"quantity"`
	Note                *string   `json:"note"`
	CreatedAt           string    `json:"created_at"`
	UpdatedAt           string    `json:"updated_at"`
	CreatedBy           uuid.UUID `json:"created_by"`
	UpdatedBy           uuid.UUID `json:"updated_by"`
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

	createdFostering, err := h.createPigletFostering.Execute(r.Context(), pigletfosteringapplication.CreatePigletFosteringCommand{
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

	platformhttp.WriteJSON(w, http.StatusCreated, toPigletFosteringResponse(createdFostering))
}

// list handles GET /api/v1/piglet-fosterings and returns the fosterings matching the filter.
// It parses the optional donor_sow_id, receiver_sow_id, from and to query parameters.
func (h *PigletFosteringHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parsePigletFosteringFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	fosterings, err := h.listPigletFosterings.Execute(r.Context(), filter)
	if err != nil {
		writePigletFosteringError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toPigletFosteringResponses(fosterings))
}

// parsePigletFosteringFilter reads the optional list filters from the query string.
// It returns an error for a malformed sow or date value and rejects a date range
// whose start is later than its end.
func parsePigletFosteringFilter(r *http.Request) (ports.PigletFosteringFilter, error) {
	query := r.URL.Query()
	var filter ports.PigletFosteringFilter

	if donor := strings.TrimSpace(query.Get("donor_sow_id")); donor != "" {
		donorSowID, err := uuid.Parse(donor)
		if err != nil {
			return ports.PigletFosteringFilter{}, errors.New("El identificador de la cerda donante no es válido")
		}
		filter.DonorSowID = &donorSowID
	}

	if receiver := strings.TrimSpace(query.Get("receiver_sow_id")); receiver != "" {
		receiverSowID, err := uuid.Parse(receiver)
		if err != nil {
			return ports.PigletFosteringFilter{}, errors.New("El identificador de la cerda receptora no es válido")
		}
		filter.ReceiverSowID = &receiverSowID
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

// toPigletFosteringResponse maps a domain fostering to the HTTP response shape.
// It formats the movement date as YYYY-MM-DD and the timestamps in UTC.
func toPigletFosteringResponse(fostering *pigletfosteringdomain.PigletFostering) pigletFosteringResponse {
	return pigletFosteringResponse{
		ID:                  fostering.ID,
		DonorFarrowingID:    fostering.DonorFarrowingID,
		ReceiverFarrowingID: fostering.ReceiverFarrowingID,
		DonorSowID:          fostering.DonorSowID,
		ReceiverSowID:       fostering.ReceiverSowID,
		MovementDate:        fostering.MovementDate.UTC().Format(dateLayout),
		Quantity:            fostering.Quantity,
		Note:                fostering.Note,
		CreatedAt:           fostering.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:           fostering.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		CreatedBy:           fostering.CreatedBy,
		UpdatedBy:           fostering.UpdatedBy,
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
		errors.Is(err, farrowingdomain.ErrInsufficientPiglets):
		status = http.StatusConflict
	case errors.Is(err, pigletfosteringdomain.ErrInvalidID),
		errors.Is(err, pigletfosteringdomain.ErrInvalidDonor),
		errors.Is(err, pigletfosteringdomain.ErrInvalidReceiver),
		errors.Is(err, pigletfosteringdomain.ErrInvalidDonorSow),
		errors.Is(err, pigletfosteringdomain.ErrInvalidReceiverSow),
		errors.Is(err, pigletfosteringdomain.ErrInvalidMovementDate),
		errors.Is(err, pigletfosteringdomain.ErrInvalidFarrowDate),
		errors.Is(err, pigletfosteringdomain.ErrInvalidQuantity),
		errors.Is(err, pigletfosteringdomain.ErrInvalidNote),
		errors.Is(err, pigletfosteringdomain.ErrInvalidCreatedBy),
		errors.Is(err, pigletfosteringdomain.ErrInvalidUpdatedBy),
		errors.Is(err, pigletfosteringdomain.ErrSameFarrowing),
		errors.Is(err, pigletfosteringdomain.ErrMovementDateBeforeFarrowing),
		errors.Is(err, pigletfosteringdomain.ErrMovementDateInFuture),
		errors.Is(err, farrowingdomain.ErrInvalidPigletQuantity):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
