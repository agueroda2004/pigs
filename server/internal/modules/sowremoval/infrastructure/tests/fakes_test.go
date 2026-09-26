package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	sowremovalapplication "server/internal/modules/sowremoval/application"
	sowremovaldomain "server/internal/modules/sowremoval/domain"
	sowremovalinfra "server/internal/modules/sowremoval/infrastructure"
	"server/internal/modules/sowremoval/ports"
	userdomain "server/internal/modules/user/domain"
)

type fakeCreateSowRemovalUseCase struct {
	removal *sowremovaldomain.SowRemoval
	err     error
	command sowremovalapplication.CreateSowRemovalCommand
	called  bool
}

func (f *fakeCreateSowRemovalUseCase) Execute(_ context.Context, command sowremovalapplication.CreateSowRemovalCommand) (*sowremovaldomain.SowRemoval, error) {
	f.called = true
	f.command = command
	return f.removal, f.err
}

type fakeListSowRemovalsUseCase struct {
	removals []*sowremovaldomain.SowRemoval
	err      error
	filter   ports.SowRemovalFilter
	called   bool
}

func (f *fakeListSowRemovalsUseCase) Execute(_ context.Context, filter ports.SowRemovalFilter) ([]*sowremovaldomain.SowRemoval, error) {
	f.called = true
	f.filter = filter
	return f.removals, f.err
}

func newTestHandler(create sowremovalinfra.CreateSowRemovalUseCase, list sowremovalinfra.ListSowRemovalsUseCase) *sowremovalinfra.SowRemovalHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return sowremovalinfra.NewSowRemovalHandler(create, list, passThrough, passThrough)
}

func authenticatedRequest(request *http.Request, userID uuid.UUID) *http.Request {
	return request.WithContext(authdomain.WithAuthenticatedUser(request.Context(), authdomain.AuthenticatedUser{
		UserID: userID,
		Role:   userdomain.RoleAdmin,
	}))
}

func serve(handler *sowremovalinfra.SowRemovalHandler, request *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func handlerRemoval() *sowremovaldomain.SowRemoval {
	actor := uuid.New()
	createdAt := time.Date(2026, time.January, 20, 3, 4, 5, 0, time.UTC)
	return &sowremovaldomain.SowRemoval{
		ID:          uuid.New(),
		SowID:       uuid.New(),
		RemovalDate: time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC),
		Type:        sowremovaldomain.TypeDeath,
		Reason:      sowremovaldomain.ReasonDisease,
		LastState:   "Gestando",
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
		CreatedBy:   actor,
		UpdatedBy:   actor,
	}
}
