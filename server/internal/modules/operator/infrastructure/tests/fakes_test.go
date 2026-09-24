package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	operatorapplication "server/internal/modules/operator/application"
	operatordomain "server/internal/modules/operator/domain"
	operatorinfra "server/internal/modules/operator/infrastructure"
	userdomain "server/internal/modules/user/domain"
)

type fakeCreateOperatorUseCase struct {
	operator *operatordomain.Operator
	err      error
	command  operatorapplication.CreateOperatorCommand
	called   bool
}

func (f *fakeCreateOperatorUseCase) Execute(_ context.Context, command operatorapplication.CreateOperatorCommand) (*operatordomain.Operator, error) {
	f.called = true
	f.command = command
	return f.operator, f.err
}

type fakeListOperatorsUseCase struct {
	operators []*operatordomain.Operator
	err       error
	called    bool
}

func (f *fakeListOperatorsUseCase) Execute(_ context.Context) ([]*operatordomain.Operator, error) {
	f.called = true
	return f.operators, f.err
}

type fakeUpdateOperatorUseCase struct {
	operator   *operatordomain.Operator
	err        error
	operatorID uuid.UUID
	command    operatorapplication.UpdateOperatorCommand
	called     bool
}

func (f *fakeUpdateOperatorUseCase) Execute(_ context.Context, operatorID uuid.UUID, command operatorapplication.UpdateOperatorCommand) (*operatordomain.Operator, error) {
	f.called = true
	f.operatorID = operatorID
	f.command = command
	return f.operator, f.err
}

func newTestHandler(create operatorinfra.CreateOperatorUseCase, list operatorinfra.ListOperatorsUseCase, update operatorinfra.UpdateOperatorUseCase) *operatorinfra.OperatorHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return operatorinfra.NewOperatorHandler(create, list, update, passThrough, passThrough)
}

func authenticatedRequest(request *http.Request, userID uuid.UUID) *http.Request {
	return request.WithContext(authdomain.WithAuthenticatedUser(request.Context(), authdomain.AuthenticatedUser{
		UserID: userID,
		Role:   userdomain.RoleAdmin,
	}))
}

func serve(handler *operatorinfra.OperatorHandler, request *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func handlerOperator() *operatordomain.Operator {
	actor := uuid.New()
	return &operatordomain.Operator{
		ID: uuid.New(), Name: "Juan Pérez", Active: true,
		CreatedAt: time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC),
		UpdatedAt: time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC),
		CreatedBy: actor, UpdatedBy: actor,
	}
}
