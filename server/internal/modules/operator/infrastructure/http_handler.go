package infrastructure

import (
	"context"
	"errors"
	"net/http"

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
	Execute(context.Context) ([]*operatordomain.Operator, error)
}

type UpdateOperatorUseCase interface {
	Execute(context.Context, uuid.UUID, operatorapplication.UpdateOperatorCommand) (*operatordomain.Operator, error)
}

type OperatorHandler struct {
	createOperator  CreateOperatorUseCase
	listOperators   ListOperatorsUseCase
	updateOperator  UpdateOperatorUseCase
	authMiddleware  func(http.Handler) http.Handler
	adminMiddleware func(http.Handler) http.Handler
}

// NewOperatorHandler wires the operator use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewOperatorHandler(
	createOperator CreateOperatorUseCase,
	listOperators ListOperatorsUseCase,
	updateOperator UpdateOperatorUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *OperatorHandler {
	return &OperatorHandler{
		createOperator:  createOperator,
		listOperators:   listOperators,
		updateOperator:  updateOperator,
		authMiddleware:  authMiddleware,
		adminMiddleware: adminMiddleware,
	}
}

// RegisterRoutes registers the operator create, list and update endpoints on the mux.
// Write routes are admin-only while the list route only requires authentication.
func (h *OperatorHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/operators", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/operators", h.authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("PATCH /api/v1/operators/{id}", h.adminMiddleware(http.HandlerFunc(h.update)))
}

type createOperatorRequest struct {
	Name string `json:"name"`
}

type updateOperatorRequest struct {
	Name   *string `json:"name"`
	Active *bool   `json:"active"`
}

type operatorResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	CreatedBy uuid.UUID `json:"created_by"`
	UpdatedBy uuid.UUID `json:"updated_by"`
}

// create handles POST /api/v1/operators and creates an operator from the request body.
// It reads the actor from the context to set CreatedBy.
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

	createdOperator, err := h.createOperator.Execute(r.Context(), operatorapplication.CreateOperatorCommand{
		Name:      request.Name,
		CreatedBy: actor.UserID,
	})
	if err != nil {
		writeOperatorError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusCreated, toOperatorResponse(createdOperator))
}

// list handles GET /api/v1/operators and returns every registered operator.
// It maps the domain operators to the public response shape.
func (h *OperatorHandler) list(w http.ResponseWriter, r *http.Request) {
	operators, err := h.listOperators.Execute(r.Context())
	if err != nil {
		writeOperatorError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toOperatorResponses(operators))
}

// update handles PATCH /api/v1/operators/{id} and applies the provided fields.
// It parses the id path value and reads the actor to set UpdatedBy.
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

	updatedOperator, err := h.updateOperator.Execute(r.Context(), operatorID, operatorapplication.UpdateOperatorCommand{
		Name:      request.Name,
		Active:    request.Active,
		UpdatedBy: actor.UserID,
	})
	if err != nil {
		writeOperatorError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toOperatorResponse(updatedOperator))
}

// toOperatorResponse maps a domain operator to the HTTP response shape.
// It formats timestamps in UTC.
func toOperatorResponse(operator *operatordomain.Operator) operatorResponse {
	return operatorResponse{
		ID:        operator.ID,
		Name:      operator.Name,
		Active:    operator.Active,
		CreatedAt: operator.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt: operator.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		CreatedBy: operator.CreatedBy,
		UpdatedBy: operator.UpdatedBy,
	}
}

// toOperatorResponses maps a list of domain operators to HTTP response shapes.
// It returns an empty slice instead of null when there are no operators.
func toOperatorResponses(operators []*operatordomain.Operator) []operatorResponse {
	responses := make([]operatorResponse, 0, len(operators))
	for _, operator := range operators {
		responses = append(responses, toOperatorResponse(operator))
	}
	return responses
}

// writeOperatorError maps operator domain and port errors to HTTP status codes.
// It falls back to 500 for unrecognized errors.
func writeOperatorError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrOperatorNameAlreadyUsed):
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
