package infrastructure

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	serviceapplication "server/internal/modules/service/application"
	servicedomain "server/internal/modules/service/domain"
	"server/internal/modules/service/ports"
	platformhttp "server/internal/platform/http"
)

const dateLayout = "2006-01-02"

type CreateServiceUseCase interface {
	Execute(context.Context, serviceapplication.CreateServiceCommand) (*servicedomain.Service, error)
}

type ListServicesUseCase interface {
	Execute(context.Context, ports.ServiceFilter) ([]*servicedomain.Service, error)
}

type ServiceHandler struct {
	createService   CreateServiceUseCase
	listServices    ListServicesUseCase
	authMiddleware  func(http.Handler) http.Handler
	adminMiddleware func(http.Handler) http.Handler
}

// NewServiceHandler wires the service use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewServiceHandler(
	createService CreateServiceUseCase,
	listServices ListServicesUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *ServiceHandler {
	return &ServiceHandler{
		createService:   createService,
		listServices:    listServices,
		authMiddleware:  authMiddleware,
		adminMiddleware: adminMiddleware,
	}
}

// RegisterRoutes registers the service create and list endpoints on the mux.
// Creation is admin-only while the list route only requires authentication.
func (h *ServiceHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/services", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/services", h.authMiddleware(http.HandlerFunc(h.list)))
}

type createServiceRequest struct {
	SowID    string               `json:"sow_id"`
	Note     *string              `json:"note"`
	Location *string              `json:"location"`
	Mounts   []createMountRequest `json:"mounts"`
}

type createMountRequest struct {
	BoarID     string  `json:"boar_id"`
	OperatorID string  `json:"operator_id"`
	MountDate  string  `json:"mount_date"`
	Type       string  `json:"type"`
	Note       *string `json:"note"`
}

type serviceResponse struct {
	ID                    uuid.UUID       `json:"id"`
	SowID                 uuid.UUID       `json:"sow_id"`
	ExpectedFarrowingDate *string         `json:"expected_farrowing_date"`
	Note                  *string         `json:"note"`
	State                 string          `json:"state"`
	Location              *string         `json:"location"`
	Mounts                []mountResponse `json:"mounts"`
	CreatedAt             string          `json:"created_at"`
	UpdatedAt             string          `json:"updated_at"`
	CreatedBy             uuid.UUID       `json:"created_by"`
	UpdatedBy             uuid.UUID       `json:"updated_by"`
}

type mountResponse struct {
	ID          uuid.UUID `json:"id"`
	ServiceID   uuid.UUID `json:"service_id"`
	BoarID      uuid.UUID `json:"boar_id"`
	OperatorID  uuid.UUID `json:"operator_id"`
	MountNumber int       `json:"mount_number"`
	MountDate   string    `json:"mount_date"`
	Type        string    `json:"type"`
	Note        *string   `json:"note"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
	CreatedBy   uuid.UUID `json:"created_by"`
	UpdatedBy   uuid.UUID `json:"updated_by"`
}

// create handles POST /api/v1/services and registers a service with its mounts.
// It parses the mount dates, reads the actor from the context and never accepts
// a state or an expected farrowing date, which are derived by the domain.
func (h *ServiceHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createServiceRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	sowID, err := uuid.Parse(request.SowID)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La cerda no es válida"))
		return
	}

	mounts := make([]serviceapplication.CreateMountCommand, 0, len(request.Mounts))
	for _, mount := range request.Mounts {
		boarID, err := uuid.Parse(mount.BoarID)
		if err != nil {
			platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El verraco no es válido"))
			return
		}

		operatorID, err := uuid.Parse(mount.OperatorID)
		if err != nil {
			platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El operador no es válido"))
			return
		}

		mountDate, err := time.Parse(dateLayout, mount.MountDate)
		if err != nil {
			platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La fecha de monta no es válida"))
			return
		}

		mounts = append(mounts, serviceapplication.CreateMountCommand{
			BoarID:     boarID,
			OperatorID: operatorID,
			MountDate:  mountDate,
			Type:       servicedomain.MountType(strings.TrimSpace(mount.Type)),
			Note:       mount.Note,
		})
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	createdService, err := h.createService.Execute(r.Context(), serviceapplication.CreateServiceCommand{
		SowID:     sowID,
		Note:      request.Note,
		Location:  request.Location,
		Mounts:    mounts,
		CreatedBy: actor.UserID,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusCreated, toServiceResponse(createdService))
}

// list handles GET /api/v1/services and returns the services matching the filter.
// It parses the optional sow_id and state query parameters.
func (h *ServiceHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseServiceFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	services, err := h.listServices.Execute(r.Context(), filter)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toServiceResponses(services))
}

// parseServiceFilter reads the optional list filters from the query string.
// It returns an error for a malformed sow_id or state value.
func parseServiceFilter(r *http.Request) (ports.ServiceFilter, error) {
	query := r.URL.Query()
	var filter ports.ServiceFilter

	if sow := strings.TrimSpace(query.Get("sow_id")); sow != "" {
		sowID, err := uuid.Parse(sow)
		if err != nil {
			return ports.ServiceFilter{}, errors.New("El identificador de la cerda no es válido")
		}
		filter.SowID = &sowID
	}

	if state := strings.TrimSpace(query.Get("state")); state != "" {
		parsedState, err := servicedomain.ParseState(state)
		if err != nil {
			return ports.ServiceFilter{}, errors.New("El estado no es válido")
		}
		filter.State = &parsedState
	}

	return filter, nil
}

// toServiceResponse maps a domain service to the HTTP response shape.
// It formats dates as YYYY-MM-DD and timestamps in UTC.
func toServiceResponse(service *servicedomain.Service) serviceResponse {
	return serviceResponse{
		ID:                    service.ID,
		SowID:                 service.SowID,
		ExpectedFarrowingDate: formatOptionalDate(service.ExpectedFarrowingDate),
		Note:                  service.Note,
		State:                 string(service.State),
		Location:              service.Location,
		Mounts:                toMountResponses(service.Mounts),
		CreatedAt:             service.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:             service.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		CreatedBy:             service.CreatedBy,
		UpdatedBy:             service.UpdatedBy,
	}
}

// toServiceResponses maps a list of domain services to HTTP response shapes.
// It returns an empty slice instead of null when there are no services.
func toServiceResponses(services []*servicedomain.Service) []serviceResponse {
	responses := make([]serviceResponse, 0, len(services))
	for _, service := range services {
		responses = append(responses, toServiceResponse(service))
	}
	return responses
}

// toMountResponses maps the mounts of a service to their response shapes.
// It returns an empty slice instead of null when the service has no mounts.
func toMountResponses(mounts []*servicedomain.Mount) []mountResponse {
	responses := make([]mountResponse, 0, len(mounts))
	for _, mount := range mounts {
		responses = append(responses, mountResponse{
			ID:          mount.ID,
			ServiceID:   mount.ServiceID,
			BoarID:      mount.BoarID,
			OperatorID:  mount.OperatorID,
			MountNumber: mount.MountNumber,
			MountDate:   mount.MountDate.UTC().Format(dateLayout),
			Type:        string(mount.Type),
			Note:        mount.Note,
			CreatedAt:   mount.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
			UpdatedAt:   mount.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
			CreatedBy:   mount.CreatedBy,
			UpdatedBy:   mount.UpdatedBy,
		})
	}
	return responses
}

// formatOptionalDate formats an optional date as YYYY-MM-DD.
// It returns nil when the date pointer is nil.
func formatOptionalDate(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(dateLayout)
	return &formatted
}

// writeServiceError maps service domain, application and port errors to status codes.
// It falls back to 500 for unrecognized errors.
func writeServiceError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrServiceNotFound),
		errors.Is(err, ports.ErrSowNotFound),
		errors.Is(err, ports.ErrBoarNotFound),
		errors.Is(err, ports.ErrOperatorNotFound):
		status = http.StatusNotFound
	case errors.Is(err, serviceapplication.ErrSowNotEligible),
		errors.Is(err, serviceapplication.ErrBoarNotEligible),
		errors.Is(err, serviceapplication.ErrOperatorNotAvailable):
		status = http.StatusConflict
	case errors.Is(err, servicedomain.ErrInvalidID),
		errors.Is(err, servicedomain.ErrInvalidSow),
		errors.Is(err, servicedomain.ErrInvalidLocation),
		errors.Is(err, servicedomain.ErrInvalidNote),
		errors.Is(err, servicedomain.ErrInvalidState),
		errors.Is(err, servicedomain.ErrInvalidCreatedBy),
		errors.Is(err, servicedomain.ErrInvalidUpdatedBy),
		errors.Is(err, servicedomain.ErrInvalidMounts),
		errors.Is(err, servicedomain.ErrMountDatesNotAscending),
		errors.Is(err, servicedomain.ErrMountDateGapTooLarge),
		errors.Is(err, servicedomain.ErrMountDateInFuture),
		errors.Is(err, servicedomain.ErrInvalidMountID),
		errors.Is(err, servicedomain.ErrInvalidMountService),
		errors.Is(err, servicedomain.ErrInvalidMountBoar),
		errors.Is(err, servicedomain.ErrInvalidMountOperator),
		errors.Is(err, servicedomain.ErrInvalidMountNumber),
		errors.Is(err, servicedomain.ErrInvalidMountDate),
		errors.Is(err, servicedomain.ErrInvalidMountType),
		errors.Is(err, servicedomain.ErrInvalidMountNote),
		errors.Is(err, servicedomain.ErrInvalidMountCreatedBy):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
