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
	Execute(context.Context, ports.MedicationFilter) ([]medicationdomain.MedicationListItem, error)
}

type ListMedicationOptionsUseCase interface {
	Execute(context.Context, *bool) ([]medicationdomain.MedicationOption, error)
}

type UpdateMedicationUseCase interface {
	Execute(context.Context, uuid.UUID, medicationapplication.UpdateMedicationCommand) (*medicationdomain.Medication, error)
}

type DeleteMedicationUseCase interface {
	Execute(context.Context, uuid.UUID) error
}

type MedicationHandler struct {
	createMedication      CreateMedicationUseCase
	listMedications       ListMedicationsUseCase
	listMedicationOptions ListMedicationOptionsUseCase
	updateMedication      UpdateMedicationUseCase
	deleteMedication      DeleteMedicationUseCase
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
	deleteMedication DeleteMedicationUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *MedicationHandler {
	return &MedicationHandler{
		createMedication:      createMedication,
		listMedications:       listMedications,
		listMedicationOptions: listMedicationOptions,
		updateMedication:      updateMedication,
		deleteMedication:      deleteMedication,
		authMiddleware:        authMiddleware,
		adminMiddleware:       adminMiddleware,
	}
}

// RegisterRoutes registers the medication create, list, options, update and delete endpoints.
// Write routes are admin-only while the read routes only require authentication.
func (h *MedicationHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/medications", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/medications", h.authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("GET /api/v1/medications/options", h.authMiddleware(http.HandlerFunc(h.listOptions)))
	mux.Handle("PATCH /api/v1/medications/{id}", h.adminMiddleware(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /api/v1/medications/{id}", h.adminMiddleware(http.HandlerFunc(h.delete)))
}

type createMedicationRequest struct {
	Name string `json:"name"`
}

type updateMedicationRequest struct {
	Name   *string `json:"name"`
	Active *bool   `json:"active"`
}

type medicationListResponse struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Active bool      `json:"active"`
}

type medicationOptionResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// create handles POST /api/v1/medications and creates a medication from the request body.
// It reads the actor from the context to set CreatedBy and answers with a status code only.
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

	if _, err := h.createMedication.Execute(r.Context(), medicationapplication.CreateMedicationCommand{
		Name:      request.Name,
		CreatedBy: actor.UserID,
	}); err != nil {
		writeMedicationError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
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

	platformhttp.WriteJSON(w, http.StatusOK, toMedicationListResponses(medications))
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
// It rejects a name longer than MaxNameLength and a malformed active value.
func parseMedicationFilter(r *http.Request) (ports.MedicationFilter, error) {
	query := r.URL.Query()
	var filter ports.MedicationFilter

	if name := strings.TrimSpace(query.Get("name")); name != "" {
		if len([]rune(name)) > medicationdomain.MaxNameLength {
			return ports.MedicationFilter{}, medicationdomain.ErrInvalidName
		}
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
// It parses the id path value, reads the actor to set UpdatedBy and answers with a status code only.
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

	if _, err := h.updateMedication.Execute(r.Context(), medicationID, medicationapplication.UpdateMedicationCommand{
		Name:      request.Name,
		Active:    request.Active,
		UpdatedBy: actor.UserID,
	}); err != nil {
		writeMedicationError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// delete handles DELETE /api/v1/medications/{id} and removes a medication.
// It parses the id path value, maps a linked-records conflict to 409 and
// responds with 204 and no body.
func (h *MedicationHandler) delete(w http.ResponseWriter, r *http.Request) {
	medicationID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del medicamento no es válido"))
		return
	}

	if err := h.deleteMedication.Execute(r.Context(), medicationID); err != nil {
		writeMedicationError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// toMedicationListResponses maps medication read models to their list response shape.
// It returns an empty slice instead of null when there are no medications.
func toMedicationListResponses(medications []medicationdomain.MedicationListItem) []medicationListResponse {
	responses := make([]medicationListResponse, 0, len(medications))
	for _, medication := range medications {
		responses = append(responses, medicationListResponse{
			ID:     medication.ID,
			Name:   medication.Name,
			Active: medication.Active,
		})
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
	case errors.Is(err, ports.ErrMedicationNameAlreadyUsed),
		errors.Is(err, ports.ErrMedicationInUse):
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
