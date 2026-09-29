package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	boarremovalapplication "server/internal/modules/boarremoval/application"
	boarremovaldomain "server/internal/modules/boarremoval/domain"
	boarremovalinfra "server/internal/modules/boarremoval/infrastructure"
	"server/internal/modules/boarremoval/ports"
	userdomain "server/internal/modules/user/domain"
)

type fakeCreateBoarRemovalUseCase struct {
	removal *boarremovaldomain.BoarRemoval
	err     error
	command boarremovalapplication.CreateBoarRemovalCommand
	called  bool
}

func (f *fakeCreateBoarRemovalUseCase) Execute(_ context.Context, command boarremovalapplication.CreateBoarRemovalCommand) (*boarremovaldomain.BoarRemoval, error) {
	f.called = true
	f.command = command
	return f.removal, f.err
}

type fakeListBoarRemovalsUseCase struct {
	removals []*boarremovaldomain.BoarRemoval
	err      error
	filter   ports.BoarRemovalFilter
	called   bool
}

func (f *fakeListBoarRemovalsUseCase) Execute(_ context.Context, filter ports.BoarRemovalFilter) ([]*boarremovaldomain.BoarRemoval, error) {
	f.called = true
	f.filter = filter
	return f.removals, f.err
}

type fakeUpdateBoarRemovalUseCase struct {
	removal *boarremovaldomain.BoarRemoval
	err     error
	id      uuid.UUID
	command boarremovalapplication.UpdateBoarRemovalCommand
	called  bool
}

func (f *fakeUpdateBoarRemovalUseCase) Execute(_ context.Context, id uuid.UUID, command boarremovalapplication.UpdateBoarRemovalCommand) (*boarremovaldomain.BoarRemoval, error) {
	f.called = true
	f.id = id
	f.command = command
	return f.removal, f.err
}

type fakeDeleteBoarRemovalUseCase struct {
	err     error
	command boarremovalapplication.DeleteBoarRemovalCommand
	called  bool
}

func (f *fakeDeleteBoarRemovalUseCase) Execute(_ context.Context, command boarremovalapplication.DeleteBoarRemovalCommand) error {
	f.called = true
	f.command = command
	return f.err
}

func newTestHandler(create boarremovalinfra.CreateBoarRemovalUseCase, list boarremovalinfra.ListBoarRemovalsUseCase) *boarremovalinfra.BoarRemovalHandler {
	return newTestHandlerWithUpdate(create, &fakeUpdateBoarRemovalUseCase{}, list)
}

func newTestHandlerWithUpdate(
	create boarremovalinfra.CreateBoarRemovalUseCase,
	update boarremovalinfra.UpdateBoarRemovalUseCase,
	list boarremovalinfra.ListBoarRemovalsUseCase,
) *boarremovalinfra.BoarRemovalHandler {
	return newTestHandlerWithDelete(create, update, &fakeDeleteBoarRemovalUseCase{}, list)
}

func newTestHandlerWithDelete(
	create boarremovalinfra.CreateBoarRemovalUseCase,
	update boarremovalinfra.UpdateBoarRemovalUseCase,
	deleteUseCase boarremovalinfra.DeleteBoarRemovalUseCase,
	list boarremovalinfra.ListBoarRemovalsUseCase,
) *boarremovalinfra.BoarRemovalHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return boarremovalinfra.NewBoarRemovalHandler(create, update, deleteUseCase, list, passThrough, passThrough)
}

func authenticatedRequest(request *http.Request, userID uuid.UUID) *http.Request {
	return request.WithContext(authdomain.WithAuthenticatedUser(request.Context(), authdomain.AuthenticatedUser{
		UserID: userID,
		Role:   userdomain.RoleAdmin,
	}))
}

func serve(handler *boarremovalinfra.BoarRemovalHandler, request *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func handlerRemoval() *boarremovaldomain.BoarRemoval {
	actor := uuid.New()
	createdAt := time.Date(2026, time.January, 20, 3, 4, 5, 0, time.UTC)
	return &boarremovaldomain.BoarRemoval{
		ID:          uuid.New(),
		BoarID:      uuid.New(),
		RemovalDate: time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC),
		Type:        boarremovaldomain.TypeDeath,
		Reason:      boarremovaldomain.ReasonDisease,
		LastState:   "Vivo",
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
		CreatedBy:   actor,
		UpdatedBy:   actor,
	}
}
