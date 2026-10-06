package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	partialweagingapplication "server/internal/modules/partialweaging/application"
	partialweagingdomain "server/internal/modules/partialweaging/domain"
	partialweaginginfra "server/internal/modules/partialweaging/infrastructure"
	"server/internal/modules/partialweaging/ports"
	userdomain "server/internal/modules/user/domain"
)

type fakeCreatePartialWeagingUseCase struct {
	weaging *partialweagingdomain.PartialWeaging
	err     error
	command partialweagingapplication.CreatePartialWeagingCommand
	called  bool
}

func (f *fakeCreatePartialWeagingUseCase) Execute(_ context.Context, command partialweagingapplication.CreatePartialWeagingCommand) (*partialweagingdomain.PartialWeaging, error) {
	f.called = true
	f.command = command
	return f.weaging, f.err
}

type fakeListPartialWeagingsUseCase struct {
	weagings []*partialweagingdomain.PartialWeaging
	err      error
	filter   ports.PartialWeagingFilter
	called   bool
}

func (f *fakeListPartialWeagingsUseCase) Execute(_ context.Context, filter ports.PartialWeagingFilter) ([]*partialweagingdomain.PartialWeaging, error) {
	f.called = true
	f.filter = filter
	return f.weagings, f.err
}

func newTestHandler(create partialweaginginfra.CreatePartialWeagingUseCase, list partialweaginginfra.ListPartialWeagingsUseCase) *partialweaginginfra.PartialWeagingHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return partialweaginginfra.NewPartialWeagingHandler(create, list, passThrough, passThrough)
}

func newAdminGuardHandler(create partialweaginginfra.CreatePartialWeagingUseCase) *partialweaginginfra.PartialWeagingHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	adminGuard := func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
	}
	return partialweaginginfra.NewPartialWeagingHandler(create, &fakeListPartialWeagingsUseCase{}, passThrough, adminGuard)
}

func authenticatedRequest(request *http.Request, userID uuid.UUID) *http.Request {
	return request.WithContext(authdomain.WithAuthenticatedUser(request.Context(), authdomain.AuthenticatedUser{
		UserID: userID,
		Role:   userdomain.RoleAdmin,
	}))
}

func serve(handler *partialweaginginfra.PartialWeagingHandler, request *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func handlerPartialWeaging() *partialweagingdomain.PartialWeaging {
	actor := uuid.New()
	createdAt := time.Date(2026, time.April, 25, 3, 4, 5, 0, time.UTC)
	weight := 42.5
	note := "Destete parcial de la camada"
	return &partialweagingdomain.PartialWeaging{
		ID:          uuid.New(),
		FarrowingID: uuid.New(),
		SowID:       uuid.New(),
		WeagingDate: time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC),
		Quantity:    2,
		TotalWeight: &weight,
		Type:        partialweagingdomain.TypeNormal,
		Note:        &note,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
		CreatedBy:   actor,
		UpdatedBy:   actor,
	}
}
