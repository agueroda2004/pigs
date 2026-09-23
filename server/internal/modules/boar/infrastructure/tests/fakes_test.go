package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	boarapplication "server/internal/modules/boar/application"
	boardomain "server/internal/modules/boar/domain"
	boarinfra "server/internal/modules/boar/infrastructure"
	"server/internal/modules/boar/ports"
	userdomain "server/internal/modules/user/domain"
)

type fakeCreateBoarUseCase struct {
	boar    *boardomain.Boar
	err     error
	command boarapplication.CreateBoarCommand
	called  bool
}

func (f *fakeCreateBoarUseCase) Execute(_ context.Context, command boarapplication.CreateBoarCommand) (*boardomain.Boar, error) {
	f.called = true
	f.command = command
	return f.boar, f.err
}

type fakeListBoarsUseCase struct {
	boars  []*boardomain.Boar
	err    error
	filter ports.BoarFilter
	called bool
}

func (f *fakeListBoarsUseCase) Execute(_ context.Context, filter ports.BoarFilter) ([]*boardomain.Boar, error) {
	f.called = true
	f.filter = filter
	return f.boars, f.err
}

type fakeUpdateBoarUseCase struct {
	boar    *boardomain.Boar
	err     error
	boarID  uuid.UUID
	command boarapplication.UpdateBoarCommand
	called  bool
}

func (f *fakeUpdateBoarUseCase) Execute(_ context.Context, boarID uuid.UUID, command boarapplication.UpdateBoarCommand) (*boardomain.Boar, error) {
	f.called = true
	f.boarID = boarID
	f.command = command
	return f.boar, f.err
}

func newTestHandler(create boarinfra.CreateBoarUseCase, list boarinfra.ListBoarsUseCase, update boarinfra.UpdateBoarUseCase) *boarinfra.BoarHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return boarinfra.NewBoarHandler(create, list, update, passThrough, passThrough)
}

func authenticatedRequest(request *http.Request, userID uuid.UUID) *http.Request {
	return request.WithContext(authdomain.WithAuthenticatedUser(request.Context(), authdomain.AuthenticatedUser{
		UserID: userID,
		Role:   userdomain.RoleAdmin,
	}))
}

func serve(handler *boarinfra.BoarHandler, request *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func handlerBoar() *boardomain.Boar {
	actor := uuid.New()
	birth := time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC)
	return &boardomain.Boar{
		ID:        uuid.New(),
		Code:      "B-001",
		Active:    true,
		EntryDate: time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC),
		BirthDate: &birth,
		State:     boardomain.StateAlive,
		Origin:    boardomain.OriginOwn,
		BreedID:   uuid.New(),
		CreatedAt: time.Date(2026, time.January, 10, 3, 4, 5, 0, time.UTC),
		UpdatedAt: time.Date(2026, time.January, 10, 3, 4, 5, 0, time.UTC),
		CreatedBy: actor,
		UpdatedBy: actor,
	}
}
