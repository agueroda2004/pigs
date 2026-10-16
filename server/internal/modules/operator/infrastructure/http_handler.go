package infrastructure

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	operatorapplication "server/internal/modules/operator/application"
	operatordomain "server/internal/modules/operator/domain"
	"server/internal/modules/operator/ports"
	platformhttp "server/internal/platform/http"
)

type CreateOperatorUseCase interface {
	Execute(context.Context, operatorapplication.CreateOperatorCommand) (*operatordomain.Operator, error)
}

type ListOperatorsUseCase interface {
	Execute(context.Context) ([]operatordomain.OperatorListItem, error)
}

type ListOperatorDropdownUseCase interface {
	Execute(context.Context, *bool) ([]operatordomain.OperatorDropdown, error)
}

type UpdateOperatorUseCase interface {
	Execute(context.Context, uuid.UUID, operatorapplication.UpdateOperatorCommand) (*operatordomain.Operator, error)
}

type DeleteOperatorUseCase interface {
	Execute(context.Context, uuid.UUID) error
}

type OperatorHandler struct {
	createOperator       CreateOperatorUseCase
	listOperators        ListOperatorsUseCase
	listOperatorDropdown ListOperatorDropdownUseCase
	updateOperator       UpdateOperatorUseCase
	deleteOperator       DeleteOperatorUseCase
	authMiddleware       func(http.Handler) http.Handler
	adminMiddleware      func(http.Handler) http.Handler
}

// NewOperatorHandler wires the operator use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewOperatorHandler(
	createOperator CreateOperatorUseCase,
	listOperators ListOperatorsUseCase,
	listOperatorDropdown ListOperatorDropdownUseCase,
	updateOperator UpdateOperatorUseCase,
	deleteOperator DeleteOperatorUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *OperatorHandler {
	return &OperatorHandler{
		createOperator:       createOperator,
		listOperators:        listOperators,
		listOperatorDropdown: listOperatorDropdown,
		updateOperator:       updateOperator,
		deleteOperator:       deleteOperator,
		authMiddleware:       authMiddleware,
		adminMiddleware:      adminMiddleware,
	}
}

// RegisterRoutes registers the operator create, list, dropdown, update and delete endpoints on the mux.
// Write routes are admin-only while the read routes only require authentication.
func (h *OperatorHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/operators", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/operators", h.authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("GET /api/v1/operators/dropdown", h.authMiddleware(http.HandlerFunc(h.listDropdown)))
	mux.Handle("PATCH /api/v1/operators/{id}", h.adminMiddleware(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /api/v1/operators/{id}", h.adminMiddleware(http.HandlerFunc(h.delete)))
}

type createOperatorRequest struct {
	Name string `json:"name"`
}

type updateOperatorRequest struct {
	Name   *string `json:"name"`
	Active *bool   `json:"active"`
}

type operatorListResponse struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Active bool      `json:"active"`
}

type operatorDropdownResponse struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Active bool      `json:"active"`
}

// create handles POST /api/v1/operators and creates an operator from the request body.
// It reads the actor from the context to set CreatedBy and answers with a status code only.
func (h *OperatorHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createOperatorRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if _, err := h.createOperator.Execute(r.Context(), operatorapplication.CreateOperatorCommand{
		Name:      request.Name,
		CreatedBy: actor.UserID,
	}); err != nil {
		writeOperatorError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// list handles GET /api/v1/operators and returns every registered operator.
// It maps the operator read models to the public list shape without audit fields.
func (h *OperatorHandler) list(w http.ResponseWriter, r *http.Request) {
	operators, err := h.listOperators.Execute(r.Context())
	if err != nil {
		writeOperatorError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toOperatorListResponses(operators))
}

// listDropdown handles GET /api/v1/operators/dropdown and returns operators for a selection list.
// It forwards the optional active filter so callers may narrow the list.
func (h *OperatorHandler) listDropdown(w http.ResponseWriter, r *http.Request) {
	active, err := parseActiveQuery(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	options, err := h.listOperatorDropdown.Execute(r.Context(), active)
	if err != nil {
		writeOperatorError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toOperatorDropdownResponses(options))
}

// parseActiveQuery reads the optional active query parameter.
// It returns nil (no filter) when absent and rejects a malformed boolean.
func parseActiveQuery(r *http.Request) (*bool, error) {
	active := strings.TrimSpace(r.URL.Query().Get("active"))
	if active == "" {
		return nil, nil
	}
	parsedActive, err := strconv.ParseBool(active)
	if err != nil {
		return nil, errors.New("El filtro de activo no es válido")
	}
	return &parsedActive, nil
}

// update handles PATCH /api/v1/operators/{id} and applies the provided fields.
// It parses the id path value, reads the actor to set UpdatedBy and answers with a status code only.
func (h *OperatorHandler) update(w http.ResponseWriter, r *http.Request) {
	operatorID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del operador no es válido"))
		return
	}

	var request updateOperatorRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if _, err := h.updateOperator.Execute(r.Context(), operatorID, operatorapplication.UpdateOperatorCommand{
		Name:      request.Name,
		Active:    request.Active,
		UpdatedBy: actor.UserID,
	}); err != nil {
		writeOperatorError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// delete handles DELETE /api/v1/operators/{id} and removes an operator.
// It parses the id path value and maps a linked-records conflict to 409.
func (h *OperatorHandler) delete(w http.ResponseWriter, r *http.Request) {
	operatorID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del operador no es válido"))
		return
	}

	if err := h.deleteOperator.Execute(r.Context(), operatorID); err != nil {
		writeOperatorError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// toOperatorListResponse maps an operator read model to the list response shape.
// It exposes the identity fields only and omits every audit field.
func toOperatorListResponse(operator operatordomain.OperatorListItem) operatorListResponse {
	return operatorListResponse{
		ID:     operator.ID,
		Name:   operator.Name,
		Active: operator.Active,
	}
}

// toOperatorListResponses maps operator read models to their list response shape.
// It returns an empty slice instead of null when there are no operators.
func toOperatorListResponses(operators []operatordomain.OperatorListItem) []operatorListResponse {
	responses := make([]operatorListResponse, 0, len(operators))
	for _, operator := range operators {
		responses = append(responses, toOperatorListResponse(operator))
	}
	return responses
}

// toOperatorDropdownResponses maps operator dropdown items to their HTTP response shape.
// It returns an empty slice instead of null when there are no options.
func toOperatorDropdownResponses(options []operatordomain.OperatorDropdown) []operatorDropdownResponse {
	responses := make([]operatorDropdownResponse, 0, len(options))
	for _, option := range options {
		responses = append(responses, operatorDropdownResponse{ID: option.ID, Name: option.Name, Active: option.Active})
	}
	return responses
}

// writeOperatorError maps operator domain and port errors to HTTP status codes.
// It falls back to 500 for unrecognized errors.
func writeOperatorError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrOperatorNameAlreadyUsed),
		errors.Is(err, ports.ErrOperatorInUse):
		status = http.StatusConflict
	case errors.Is(err, ports.ErrOperatorNotFound):
		status = http.StatusNotFound
	case errors.Is(err, operatordomain.ErrInvalidID),
		errors.Is(err, operatordomain.ErrInvalidName),
		errors.Is(err, operatordomain.ErrInvalidUpdate),
		errors.Is(err, operatordomain.ErrInvalidCreatedBy),
		errors.Is(err, operatordomain.ErrInvalidUpdatedBy):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
