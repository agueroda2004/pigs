package infrastructure

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	medicationapplication "server/internal/modules/medication/application"
	medicationdomain "server/internal/modules/medication/domain"
	"server/internal/modules/medication/ports"
	platformhttp "server/internal/platform/http"
)

type CreateMedicationUseCase interface {
	Execute(context.Context, medicationapplication.CreateMedicationCommand) (*medicationdomain.Medication, error)
}

type ListMedicationsUseCase interface {
	Execute(context.Context, ports.MedicationFilter) ([]*medicationdomain.Medication, error)
}

type ListMedicationOptionsUseCase interface {
	Execute(context.Context, *bool) ([]medicationdomain.MedicationOption, error)
}

type UpdateMedicationUseCase interface {
	Execute(context.Context, uuid.UUID, medicationapplication.UpdateMedicationCommand) (*medicationdomain.Medication, error)
}

type MedicationHandler struct {
	createMedication      CreateMedicationUseCase
	listMedications       ListMedicationsUseCase
	listMedicationOptions ListMedicationOptionsUseCase
	updateMedication      UpdateMedicationUseCase
	authMiddleware        func(http.Handler) http.Handler
	adminMiddleware       func(http.Handler) http.Handler
}

// NewMedicationHandler wires the medication use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewMedicationHandler(
	createMedication CreateMedicationUseCase,
	listMedications ListMedicationsUseCase,
	listMedicationOptions ListMedicationOptionsUseCase,
	updateMedication UpdateMedicationUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *MedicationHandler {
	return &MedicationHandler{
		createMedication:      createMedication,
		listMedications:       listMedications,
		listMedicationOptions: listMedicationOptions,
		updateMedication:      updateMedication,
		authMiddleware:        authMiddleware,
		adminMiddleware:       adminMiddleware,
	}
}

// RegisterRoutes registers the medication create, list, options and update endpoints.
// Write routes are admin-only while the read routes only require authentication.
func (h *MedicationHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/medications", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/medications", h.authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("GET /api/v1/medications/options", h.authMiddleware(http.HandlerFunc(h.listOptions)))
	mux.Handle("PATCH /api/v1/medications/{id}", h.adminMiddleware(http.HandlerFunc(h.update)))
}

type createMedicationRequest struct {
	Name string `json:"name"`
}

type updateMedicationRequest struct {
	Name   *string `json:"name"`
	Active *bool   `json:"active"`
}

type medicationResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	CreatedBy uuid.UUID `json:"created_by"`
	UpdatedBy uuid.UUID `json:"updated_by"`
}

type medicationOptionResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// create handles POST /api/v1/medications and creates a medication from the request body.
// It reads the actor from the context to set CreatedBy.
func (h *MedicationHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createMedicationRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	createdMedication, err := h.createMedication.Execute(r.Context(), medicationapplication.CreateMedicationCommand{
		Name:      request.Name,
		CreatedBy: actor.UserID,
	})
	if err != nil {
		writeMedicationError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusCreated, toMedicationResponse(createdMedication))
}

// list handles GET /api/v1/medications and returns the medications matching the query filter.
// It parses the optional name and active query parameters.
func (h *MedicationHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseMedicationFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	medications, err := h.listMedications.Execute(r.Context(), filter)
	if err != nil {
		writeMedicationError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toMedicationResponses(medications))
}

// listOptions handles GET /api/v1/medications/options and returns the dropdown options.
// It returns only active medications unless include_inactive=true is provided.
func (h *MedicationHandler) listOptions(w http.ResponseWriter, r *http.Request) {
	active, err := parseMedicationOptionsActive(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	options, err := h.listMedicationOptions.Execute(r.Context(), active)
	if err != nil {
		writeMedicationError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toMedicationOptionResponses(options))
}

// parseMedicationOptionsActive reads the optional include_inactive query parameter.
// It defaults to active-only (true) and returns nil (every medication) when included.
func parseMedicationOptionsActive(r *http.Request) (*bool, error) {
	includeInactive := strings.TrimSpace(r.URL.Query().Get("include_inactive"))
	if includeInactive == "" {
		active := true
		return &active, nil
	}
	parsed, err := strconv.ParseBool(includeInactive)
	if err != nil {
		return nil, errors.New("El filtro de inactivos no es válido")
	}
	if parsed {
		return nil, nil
	}
	active := true
	return &active, nil
}

// parseMedicationFilter reads the optional list filters from the query string.
// It returns an error for a malformed active value.
func parseMedicationFilter(r *http.Request) (ports.MedicationFilter, error) {
	query := r.URL.Query()
	var filter ports.MedicationFilter

	if name := strings.TrimSpace(query.Get("name")); name != "" {
		filter.Name = &name
	}

	if active := strings.TrimSpace(query.Get("active")); active != "" {
		parsedActive, err := strconv.ParseBool(active)
		if err != nil {
			return ports.MedicationFilter{}, errors.New("El filtro de activo no es válido")
		}
		filter.Active = &parsedActive
	}

	return filter, nil
}

// update handles PATCH /api/v1/medications/{id} and applies the provided fields.
// It parses the id path value and reads the actor to set UpdatedBy.
func (h *MedicationHandler) update(w http.ResponseWriter, r *http.Request) {
	medicationID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del medicamento no es válido"))
		return
	}

	var request updateMedicationRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	updatedMedication, err := h.updateMedication.Execute(r.Context(), medicationID, medicationapplication.UpdateMedicationCommand{
		Name:      request.Name,
		Active:    request.Active,
		UpdatedBy: actor.UserID,
	})
	if err != nil {
		writeMedicationError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toMedicationResponse(updatedMedication))
}

// toMedicationResponse maps a domain medication to the HTTP response shape.
// It formats timestamps in UTC.
func toMedicationResponse(medication *medicationdomain.Medication) medicationResponse {
	return medicationResponse{
		ID:        medication.ID,
		Name:      medication.Name,
		Active:    medication.Active,
		CreatedAt: medication.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt: medication.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		CreatedBy: medication.CreatedBy,
		UpdatedBy: medication.UpdatedBy,
	}
}

// toMedicationResponses maps a list of domain medications to HTTP response shapes.
// It returns an empty slice instead of null when there are no medications.
func toMedicationResponses(medications []*medicationdomain.Medication) []medicationResponse {
	responses := make([]medicationResponse, 0, len(medications))
	for _, medication := range medications {
		responses = append(responses, toMedicationResponse(medication))
	}
	return responses
}

// toMedicationOptionResponses maps medication options to their HTTP response shape.
// It returns an empty slice instead of null when there are no options.
func toMedicationOptionResponses(options []medicationdomain.MedicationOption) []medicationOptionResponse {
	responses := make([]medicationOptionResponse, 0, len(options))
	for _, option := range options {
		responses = append(responses, medicationOptionResponse{ID: option.ID, Name: option.Name})
	}
	return responses
}

// writeMedicationError maps medication domain and port errors to HTTP status codes.
// It falls back to 500 for unrecognized errors.
func writeMedicationError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrMedicationNameAlreadyUsed):
		status = http.StatusConflict
	case errors.Is(err, ports.ErrMedicationNotFound):
		status = http.StatusNotFound
	case errors.Is(err, medicationdomain.ErrInvalidID),
		errors.Is(err, medicationdomain.ErrInvalidName),
		errors.Is(err, medicationdomain.ErrInvalidUpdate),
		errors.Is(err, medicationdomain.ErrInvalidCreatedBy),
		errors.Is(err, medicationdomain.ErrInvalidUpdatedBy):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
