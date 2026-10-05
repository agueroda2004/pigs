package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	pigletfosteringapplication "server/internal/modules/pigletfostering/application"
	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
	pigletfosteringinfra "server/internal/modules/pigletfostering/infrastructure"
	"server/internal/modules/pigletfostering/ports"
	userdomain "server/internal/modules/user/domain"
)

type fakeCreatePigletFosteringUseCase struct {
	fostering *pigletfosteringdomain.PigletFostering
	err       error
	command   pigletfosteringapplication.CreatePigletFosteringCommand
	called    bool
}

func (f *fakeCreatePigletFosteringUseCase) Execute(_ context.Context, command pigletfosteringapplication.CreatePigletFosteringCommand) (*pigletfosteringdomain.PigletFostering, error) {
	f.called = true
	f.command = command
	return f.fostering, f.err
}

type fakeListPigletFosteringsUseCase struct {
	fosterings []*pigletfosteringdomain.PigletFostering
	err        error
	filter     ports.PigletFosteringFilter
	called     bool
}

func (f *fakeListPigletFosteringsUseCase) Execute(_ context.Context, filter ports.PigletFosteringFilter) ([]*pigletfosteringdomain.PigletFostering, error) {
	f.called = true
	f.filter = filter
	return f.fosterings, f.err
}

func newTestHandler(create pigletfosteringinfra.CreatePigletFosteringUseCase, list pigletfosteringinfra.ListPigletFosteringsUseCase) *pigletfosteringinfra.PigletFosteringHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return pigletfosteringinfra.NewPigletFosteringHandler(create, list, passThrough, passThrough)
}

func newAdminGuardHandler(create pigletfosteringinfra.CreatePigletFosteringUseCase) *pigletfosteringinfra.PigletFosteringHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	adminGuard := func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
	}
	return pigletfosteringinfra.NewPigletFosteringHandler(create, &fakeListPigletFosteringsUseCase{}, passThrough, adminGuard)
}

func authenticatedRequest(request *http.Request, userID uuid.UUID) *http.Request {
	return request.WithContext(authdomain.WithAuthenticatedUser(request.Context(), authdomain.AuthenticatedUser{
		UserID: userID,
		Role:   userdomain.RoleAdmin,
	}))
}

func serve(handler *pigletfosteringinfra.PigletFosteringHandler, request *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func handlerPigletFostering() *pigletfosteringdomain.PigletFostering {
	actor := uuid.New()
	createdAt := time.Date(2026, time.April, 25, 3, 4, 5, 0, time.UTC)
	note := "Traslado por camada numerosa"
	return &pigletfosteringdomain.PigletFostering{
		ID:                  uuid.New(),
		DonorFarrowingID:    uuid.New(),
		ReceiverFarrowingID: uuid.New(),
		DonorSowID:          uuid.New(),
		ReceiverSowID:       uuid.New(),
		MovementDate:        time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC),
		Quantity:            2,
		Note:                &note,
		CreatedAt:           createdAt,
		UpdatedAt:           createdAt,
		CreatedBy:           actor,
		UpdatedBy:           actor,
	}
}
