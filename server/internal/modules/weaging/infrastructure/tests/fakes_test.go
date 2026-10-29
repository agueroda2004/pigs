package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	userdomain "server/internal/modules/user/domain"
	weagingapplication "server/internal/modules/weaging/application"
	weagingdomain "server/internal/modules/weaging/domain"
	weaginginfra "server/internal/modules/weaging/infrastructure"
	"server/internal/modules/weaging/ports"
)

type fakeCreateWeagingUseCase struct {
	weaging *weagingdomain.Weaging
	err     error
	command weagingapplication.CreateWeagingCommand
	called  bool
}

func (f *fakeCreateWeagingUseCase) Execute(_ context.Context, command weagingapplication.CreateWeagingCommand) (*weagingdomain.Weaging, error) {
	f.called = true
	f.command = command
	return f.weaging, f.err
}

type fakeListWeagingsUseCase struct {
	page    weagingapplication.WeagingPage
	err     error
	filter  ports.WeagingFilter
	pageArg int
	called  bool
}

func (f *fakeListWeagingsUseCase) Execute(_ context.Context, filter ports.WeagingFilter, page int) (weagingapplication.WeagingPage, error) {
	f.called = true
	f.filter = filter
	f.pageArg = page
	return f.page, f.err
}

type fakeDeleteWeagingUseCase struct {
	err     error
	command weagingapplication.DeleteWeagingCommand
	called  bool
}

func (f *fakeDeleteWeagingUseCase) Execute(_ context.Context, command weagingapplication.DeleteWeagingCommand) error {
	f.called = true
	f.command = command
	return f.err
}

type fakeUpdateWeagingUseCase struct {
	err     error
	id      uuid.UUID
	command weagingapplication.UpdateWeagingCommand
	called  bool
}

func (f *fakeUpdateWeagingUseCase) Execute(_ context.Context, id uuid.UUID, command weagingapplication.UpdateWeagingCommand) error {
	f.called = true
	f.id = id
	f.command = command
	return f.err
}

func newTestHandler(create weaginginfra.CreateWeagingUseCase, list weaginginfra.ListWeagingsUseCase) *weaginginfra.WeagingHandler {
	return newTestHandlerWithUpdate(create, &fakeUpdateWeagingUseCase{}, list)
}

func newTestHandlerWithUpdate(
	create weaginginfra.CreateWeagingUseCase,
	update weaginginfra.UpdateWeagingUseCase,
	list weaginginfra.ListWeagingsUseCase,
) *weaginginfra.WeagingHandler {
	return newTestHandlerWithDelete(create, update, &fakeDeleteWeagingUseCase{}, list)
}

func newTestHandlerWithDelete(
	create weaginginfra.CreateWeagingUseCase,
	update weaginginfra.UpdateWeagingUseCase,
	remove weaginginfra.DeleteWeagingUseCase,
	list weaginginfra.ListWeagingsUseCase,
) *weaginginfra.WeagingHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return weaginginfra.NewWeagingHandler(create, list, update, remove, passThrough, passThrough)
}

func newAdminGuardHandler(create weaginginfra.CreateWeagingUseCase) *weaginginfra.WeagingHandler {
	return newAdminGuardDeleteHandlerWithCreate(create, &fakeUpdateWeagingUseCase{}, &fakeDeleteWeagingUseCase{})
}

func newAdminGuardDeleteHandler(remove weaginginfra.DeleteWeagingUseCase) *weaginginfra.WeagingHandler {
	return newAdminGuardDeleteHandlerWithCreate(&fakeCreateWeagingUseCase{}, &fakeUpdateWeagingUseCase{}, remove)
}

func newAdminGuardUpdateHandler(update weaginginfra.UpdateWeagingUseCase) *weaginginfra.WeagingHandler {
	return newAdminGuardDeleteHandlerWithCreate(&fakeCreateWeagingUseCase{}, update, &fakeDeleteWeagingUseCase{})
}

func newAdminGuardDeleteHandlerWithCreate(
	create weaginginfra.CreateWeagingUseCase,
	update weaginginfra.UpdateWeagingUseCase,
	remove weaginginfra.DeleteWeagingUseCase,
) *weaginginfra.WeagingHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	adminGuard := func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
	}
	return weaginginfra.NewWeagingHandler(create, &fakeListWeagingsUseCase{}, update, remove, passThrough, adminGuard)
}

func authenticatedRequest(request *http.Request, userID uuid.UUID) *http.Request {
	return request.WithContext(authdomain.WithAuthenticatedUser(request.Context(), authdomain.AuthenticatedUser{
		UserID: userID,
		Role:   userdomain.RoleAdmin,
	}))
}

func serve(handler *weaginginfra.WeagingHandler, request *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func handlerWeaging() *weagingdomain.Weaging {
	actor := uuid.New()
	createdAt := time.Date(2026, time.April, 25, 3, 4, 5, 0, time.UTC)
	weight := 120.5
	destination := "Nave 2"
	note := "Destete completo de la camada"
	return &weagingdomain.Weaging{
		ID:          uuid.New(),
		FarrowingID: uuid.New(),
		SowID:       uuid.New(),
		SowCode:     "C-001",
		WeagingDate: time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC),
		Quantity:    8,
		TotalWeight: &weight,
		Destination: &destination,
		Note:        &note,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
		CreatedBy:   actor,
		UpdatedBy:   actor,
	}
}
