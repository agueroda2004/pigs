package infrastructure

import (
	"net/http"
	"time"

	operatorapplication "server/internal/modules/operator/application"
	"server/internal/modules/operator/ports"
)

type Module struct {
	CreateOperator *operatorapplication.CreateOperatorService
	ListOperators  *operatorapplication.ListOperatorsService
	UpdateOperator *operatorapplication.UpdateOperatorService
	Handler        *OperatorHandler
}

// NewModule assembles the operator use cases and HTTP handler from its dependencies.
// It returns a module exposing the services and handler for wiring.
func NewModule(
	repository ports.OperatorRepository,
	clock func() time.Time,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *Module {
	createOperator := operatorapplication.NewCreateOperatorService(repository, clock)
	listOperators := operatorapplication.NewListOperatorsService(repository)
	updateOperator := operatorapplication.NewUpdateOperatorService(repository, clock)

	return &Module{
		CreateOperator: createOperator,
		ListOperators:  listOperators,
		UpdateOperator: updateOperator,
		Handler:        NewOperatorHandler(createOperator, listOperators, updateOperator, authMiddleware, adminMiddleware),
	}
}
