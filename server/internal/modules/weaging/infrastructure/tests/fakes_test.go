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
	weagings []*weagingdomain.Weaging
	err      error
	filter   ports.WeagingFilter
	called   bool
}

func (f *fakeListWeagingsUseCase) Execute(_ context.Context, filter ports.WeagingFilter) ([]*weagingdomain.Weaging, error) {
	f.called = true
	f.filter = filter
	return f.weagings, f.err
}

func newTestHandler(create weaginginfra.CreateWeagingUseCase, list weaginginfra.ListWeagingsUseCase) *weaginginfra.WeagingHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return weaginginfra.NewWeagingHandler(create, list, passThrough, passThrough)
}

func newAdminGuardHandler(create weaginginfra.CreateWeagingUseCase) *weaginginfra.WeagingHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	adminGuard := func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
	}
	return weaginginfra.NewWeagingHandler(create, &fakeListWeagingsUseCase{}, passThrough, adminGuard)
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
