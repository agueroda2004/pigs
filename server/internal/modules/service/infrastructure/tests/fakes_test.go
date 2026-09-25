package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	serviceapplication "server/internal/modules/service/application"
	servicedomain "server/internal/modules/service/domain"
	serviceinfra "server/internal/modules/service/infrastructure"
	"server/internal/modules/service/ports"
	userdomain "server/internal/modules/user/domain"
)

type fakeCreateServiceUseCase struct {
	service *servicedomain.Service
	err     error
	command serviceapplication.CreateServiceCommand
	called  bool
}

func (f *fakeCreateServiceUseCase) Execute(_ context.Context, command serviceapplication.CreateServiceCommand) (*servicedomain.Service, error) {
	f.called = true
	f.command = command
	return f.service, f.err
}

type fakeListServicesUseCase struct {
	services []*servicedomain.Service
	err      error
	filter   ports.ServiceFilter
	called   bool
}

func (f *fakeListServicesUseCase) Execute(_ context.Context, filter ports.ServiceFilter) ([]*servicedomain.Service, error) {
	f.called = true
	f.filter = filter
	return f.services, f.err
}

func newTestHandler(create serviceinfra.CreateServiceUseCase, list serviceinfra.ListServicesUseCase) *serviceinfra.ServiceHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return serviceinfra.NewServiceHandler(create, list, passThrough, passThrough)
}

func authenticatedRequest(request *http.Request, userID uuid.UUID) *http.Request {
	return request.WithContext(authdomain.WithAuthenticatedUser(request.Context(), authdomain.AuthenticatedUser{
		UserID: userID,
		Role:   userdomain.RoleAdmin,
	}))
}

func serve(handler *serviceinfra.ServiceHandler, request *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func handlerService() *servicedomain.Service {
	actor := uuid.New()
	serviceID := uuid.New()
	mountDate := time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	farrowing := mountDate.AddDate(0, 0, 114)
	createdAt := time.Date(2026, time.January, 10, 3, 4, 5, 0, time.UTC)
	return &servicedomain.Service{
		ID:                    serviceID,
		SowID:                 uuid.New(),
		ExpectedFarrowingDate: &farrowing,
		State:                 servicedomain.StateConfirmed,
		CreatedAt:             createdAt,
		UpdatedAt:             createdAt,
		CreatedBy:             actor,
		UpdatedBy:             actor,
		Mounts: []*servicedomain.Mount{{
			ID:          uuid.New(),
			ServiceID:   serviceID,
			BoarID:      uuid.New(),
			OperatorID:  uuid.New(),
			MountNumber: 1,
			MountDate:   mountDate,
			Type:        servicedomain.MountTypeArtificial,
			CreatedAt:   createdAt,
			UpdatedAt:   createdAt,
			CreatedBy:   actor,
			UpdatedBy:   actor,
		}},
	}
}
