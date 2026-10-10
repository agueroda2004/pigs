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
	serviceapplication "server/internal/modules/service/application"
	servicedomain "server/internal/modules/service/domain"
	"server/internal/modules/service/ports"
	platformhttp "server/internal/platform/http"
)

const dateLayout = "2006-01-02"

// maxSowCodeLength mirrors the length of the sows.code column.
// A longer sow code filter is rejected as invalid.
const maxSowCodeLength = 50

type CreateServiceUseCase interface {
	Execute(context.Context, serviceapplication.CreateServiceCommand) (*servicedomain.Service, error)
}

type ListServicesUseCase interface {
	Execute(context.Context, ports.ServiceFilter, int) (serviceapplication.ServicePage, error)
}

type DeleteServiceUseCase interface {
	Execute(context.Context, serviceapplication.DeleteServiceCommand) error
}

type UpdateServiceUseCase interface {
	Execute(context.Context, uuid.UUID, serviceapplication.UpdateServiceCommand) error
}

type ServiceHandler struct {
	createService   CreateServiceUseCase
	listServices    ListServicesUseCase
	deleteService   DeleteServiceUseCase
	updateService   UpdateServiceUseCase
	authMiddleware  func(http.Handler) http.Handler
	adminMiddleware func(http.Handler) http.Handler
}

// NewServiceHandler wires the service use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewServiceHandler(
	createService CreateServiceUseCase,
	listServices ListServicesUseCase,
	deleteService DeleteServiceUseCase,
	updateService UpdateServiceUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *ServiceHandler {
	return &ServiceHandler{
		createService:   createService,
		listServices:    listServices,
		deleteService:   deleteService,
		updateService:   updateService,
		authMiddleware:  authMiddleware,
		adminMiddleware: adminMiddleware,
	}
}

// RegisterRoutes registers the service create, list, update and delete endpoints.
// Write routes are admin-only while the list route only requires authentication.
func (h *ServiceHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/services", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/services", h.authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("PATCH /api/v1/services/{id}", h.adminMiddleware(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /api/v1/services/{id}", h.adminMiddleware(http.HandlerFunc(h.delete)))
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

type updateServiceRequest struct {
	Location *string                    `json:"location"`
	Note     *string                    `json:"note"`
	Mounts   updateServiceMountsRequest `json:"mounts"`
}

type updateServiceMountsRequest struct {
	Create []createMountRequest `json:"create"`
	Update []updateMountRequest `json:"update"`
	Delete []string             `json:"delete"`
}

type updateMountRequest struct {
	ID         string  `json:"id"`
	BoarID     string  `json:"boar_id"`
	OperatorID string  `json:"operator_id"`
	MountDate  string  `json:"mount_date"`
	Type       string  `json:"type"`
	Note       *string `json:"note"`
}

// serviceResponse is the service shape without the audit fields.
type serviceResponse struct {
	ID                    uuid.UUID       `json:"id"`
	SowID                 uuid.UUID       `json:"sow_id"`
	SowCode               string          `json:"sow_code"`
	ExpectedFarrowingDate *string         `json:"expected_farrowing_date"`
	Note                  *string         `json:"note"`
	State                 string          `json:"state"`
	Location              *string         `json:"location"`
	Mounts                []mountResponse `json:"mounts"`
}

// mountResponse is the mount shape without the audit fields.
type mountResponse struct {
	ID           uuid.UUID `json:"id"`
	ServiceID    uuid.UUID `json:"service_id"`
	BoarID       uuid.UUID `json:"boar_id"`
	BoarCode     string    `json:"boar_code"`
	OperatorID   uuid.UUID `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	MountNumber  int       `json:"mount_number"`
	MountDate    string    `json:"mount_date"`
	Type         string    `json:"type"`
	Note         *string   `json:"note"`
}

// servicePageResponse is one page of services together with its pagination metadata.
type servicePageResponse struct {
	Items      []serviceResponse `json:"items"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
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

	if _, err := h.createService.Execute(r.Context(), serviceapplication.CreateServiceCommand{
		SowID:     sowID,
		Note:      request.Note,
		Location:  request.Location,
		Mounts:    mounts,
		CreatedBy: actor.UserID,
	}); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// update handles PATCH /api/v1/services/{id} and applies the editable fields.
// It parses the mount operations, reads the actor and never accepts a state, a
// sow or an expected farrowing date, which are derived by the domain.
func (h *ServiceHandler) update(w http.ResponseWriter, r *http.Request) {
	serviceID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del servicio no es válido"))
		return
	}

	var request updateServiceRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	createMounts := make([]serviceapplication.CreateMountCommand, 0, len(request.Mounts.Create))
	for _, mount := range request.Mounts.Create {
		command, err := parseCreateMount(mount)
		if err != nil {
			platformhttp.WriteError(w, http.StatusBadRequest, err)
			return
		}
		createMounts = append(createMounts, command)
	}

	updateMounts := make([]serviceapplication.UpdateMountCommand, 0, len(request.Mounts.Update))
	for _, mount := range request.Mounts.Update {
		mountID, err := uuid.Parse(mount.ID)
		if err != nil {
			platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La monta no es válida"))
			return
		}
		command, err := parseCreateMount(createMountRequest{
			BoarID:     mount.BoarID,
			OperatorID: mount.OperatorID,
			MountDate:  mount.MountDate,
			Type:       mount.Type,
			Note:       mount.Note,
		})
		if err != nil {
			platformhttp.WriteError(w, http.StatusBadRequest, err)
			return
		}
		updateMounts = append(updateMounts, serviceapplication.UpdateMountCommand{
			ID:         mountID,
			BoarID:     command.BoarID,
			OperatorID: command.OperatorID,
			MountDate:  command.MountDate,
			Type:       command.Type,
			Note:       command.Note,
		})
	}

	deleteIDs := make([]uuid.UUID, 0, len(request.Mounts.Delete))
	for _, raw := range request.Mounts.Delete {
		mountID, err := uuid.Parse(raw)
		if err != nil {
			platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La monta no es válida"))
			return
		}
		deleteIDs = append(deleteIDs, mountID)
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if err := h.updateService.Execute(r.Context(), serviceID, serviceapplication.UpdateServiceCommand{
		Location:       request.Location,
		Note:           request.Note,
		CreateMounts:   createMounts,
		UpdateMounts:   updateMounts,
		DeleteMountIDs: deleteIDs,
		UpdatedBy:      actor.UserID,
	}); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// parseCreateMount parses a mount request into a create-mount command.
// It validates the boar, operator and mount date identifiers.
func parseCreateMount(mount createMountRequest) (serviceapplication.CreateMountCommand, error) {
	boarID, err := uuid.Parse(mount.BoarID)
	if err != nil {
		return serviceapplication.CreateMountCommand{}, errors.New("El verraco no es válido")
	}

	operatorID, err := uuid.Parse(mount.OperatorID)
	if err != nil {
		return serviceapplication.CreateMountCommand{}, errors.New("El operador no es válido")
	}

	mountDate, err := time.Parse(dateLayout, mount.MountDate)
	if err != nil {
		return serviceapplication.CreateMountCommand{}, errors.New("La fecha de monta no es válida")
	}

	return serviceapplication.CreateMountCommand{
		BoarID:     boarID,
		OperatorID: operatorID,
		MountDate:  mountDate,
		Type:       servicedomain.MountType(strings.TrimSpace(mount.Type)),
		Note:       mount.Note,
	}, nil
}

// list handles GET /api/v1/services and returns one page of services matching the filter.
// It parses the optional sow_code and state query parameters plus the page parameter.
func (h *ServiceHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseServiceFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	page, err := parseServicePage(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	services, err := h.listServices.Execute(r.Context(), filter, page)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toServicePageResponse(services))
}

// delete handles DELETE /api/v1/services/{id} and removes a service.
// It parses the id path value, reads the actor and maps a non-confirmed service
// or a service with related records to a 409 conflict.
func (h *ServiceHandler) delete(w http.ResponseWriter, r *http.Request) {
	serviceID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del servicio no es válido"))
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if err := h.deleteService.Execute(r.Context(), serviceapplication.DeleteServiceCommand{
		ID:        serviceID,
		DeletedBy: actor.UserID,
	}); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// parseServiceFilter reads the optional list filters from the query string.
// It returns an error for a state value or a sow code longer than 50 characters.
func parseServiceFilter(r *http.Request) (ports.ServiceFilter, error) {
	query := r.URL.Query()
	var filter ports.ServiceFilter

	if sowCode := strings.TrimSpace(query.Get("sow_code")); sowCode != "" {
		if len([]rune(sowCode)) > maxSowCodeLength {
			return ports.ServiceFilter{}, errors.New("El código de la cerda no es válido")
		}
		filter.SowCode = &sowCode
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

// parseServicePage reads the optional page query parameter.
// It defaults to page one and rejects non-positive or malformed values.
func parseServicePage(r *http.Request) (int, error) {
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

// toServiceResponse maps a domain service to the HTTP response shape.
// It formats dates as YYYY-MM-DD and omits the audit fields.
func toServiceResponse(service *servicedomain.Service) serviceResponse {
	return serviceResponse{
		ID:                    service.ID,
		SowID:                 service.SowID,
		SowCode:               service.SowCode,
		ExpectedFarrowingDate: formatOptionalDate(service.ExpectedFarrowingDate),
		Note:                  service.Note,
		State:                 string(service.State),
		Location:              service.Location,
		Mounts:                toMountResponses(service.Mounts),
	}
}

// toServicePageResponse maps one page of domain services to the paginated response shape.
// It returns an empty items slice instead of null when there are no services.
func toServicePageResponse(page serviceapplication.ServicePage) servicePageResponse {
	items := make([]serviceResponse, 0, len(page.Items))
	for _, service := range page.Items {
		items = append(items, toServiceResponse(service))
	}
	return servicePageResponse{
		Items:      items,
		Total:      page.Total,
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalPages: page.TotalPages,
	}
}

// toMountResponses maps the mounts of a service to their response shapes.
// It returns an empty slice instead of null when the service has no mounts.
func toMountResponses(mounts []*servicedomain.Mount) []mountResponse {
	responses := make([]mountResponse, 0, len(mounts))
	for _, mount := range mounts {
		responses = append(responses, mountResponse{
			ID:           mount.ID,
			ServiceID:    mount.ServiceID,
			BoarID:       mount.BoarID,
			BoarCode:     mount.BoarCode,
			OperatorID:   mount.OperatorID,
			OperatorName: mount.OperatorName,
			MountNumber:  mount.MountNumber,
			MountDate:    mount.MountDate.UTC().Format(dateLayout),
			Type:         string(mount.Type),
			Note:         mount.Note,
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
		errors.Is(err, serviceapplication.ErrOperatorNotAvailable),
		errors.Is(err, servicedomain.ErrServiceNotDeletable),
		errors.Is(err, servicedomain.ErrServiceNotEditable),
		errors.Is(err, ports.ErrServiceInUse):
		status = http.StatusConflict
	case errors.Is(err, serviceapplication.ErrMountBeforeBoarEntry),
		errors.Is(err, servicedomain.ErrInvalidID),
		errors.Is(err, servicedomain.ErrInvalidSow),
		errors.Is(err, servicedomain.ErrInvalidLocation),
		errors.Is(err, servicedomain.ErrInvalidNote),
		errors.Is(err, servicedomain.ErrInvalidState),
		errors.Is(err, servicedomain.ErrMountBeforeEntryDate),
		errors.Is(err, servicedomain.ErrMountBeforePreviousService),
		errors.Is(err, servicedomain.ErrMountBeforeAbortion),
		errors.Is(err, servicedomain.ErrMountNotFound),
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
