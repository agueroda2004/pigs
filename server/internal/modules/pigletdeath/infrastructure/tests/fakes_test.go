package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	pigletdeathapplication "server/internal/modules/pigletdeath/application"
	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	pigletdeathinfra "server/internal/modules/pigletdeath/infrastructure"
	"server/internal/modules/pigletdeath/ports"
	userdomain "server/internal/modules/user/domain"
)

type fakeCreatePigletDeathUseCase struct {
	death   *pigletdeathdomain.PigletDeath
	err     error
	command pigletdeathapplication.CreatePigletDeathCommand
	called  bool
}

func (f *fakeCreatePigletDeathUseCase) Execute(_ context.Context, command pigletdeathapplication.CreatePigletDeathCommand) (*pigletdeathdomain.PigletDeath, error) {
	f.called = true
	f.command = command
	return f.death, f.err
}

type fakeListPigletDeathsUseCase struct {
	deaths []*pigletdeathdomain.PigletDeath
	total  int
	err    error
	filter ports.PigletDeathFilter
	page   int
	called bool
}

func (f *fakeListPigletDeathsUseCase) Execute(_ context.Context, filter ports.PigletDeathFilter, page int) (pigletdeathapplication.PigletDeathPage, error) {
	f.called = true
	f.filter = filter
	f.page = page
	if f.err != nil {
		return pigletdeathapplication.PigletDeathPage{}, f.err
	}
	totalPages := 0
	if f.total > 0 {
		totalPages = (f.total + pigletdeathapplication.DefaultPigletDeathPageSize - 1) / pigletdeathapplication.DefaultPigletDeathPageSize
	}
	return pigletdeathapplication.PigletDeathPage{
		Items:      f.deaths,
		Total:      f.total,
		Page:       page,
		PageSize:   pigletdeathapplication.DefaultPigletDeathPageSize,
		TotalPages: totalPages,
	}, nil
}

type fakeDeletePigletDeathUseCase struct {
	err     error
	command pigletdeathapplication.DeletePigletDeathCommand
	called  bool
}

func (f *fakeDeletePigletDeathUseCase) Execute(_ context.Context, command pigletdeathapplication.DeletePigletDeathCommand) error {
	f.called = true
	f.command = command
	return f.err
}

type fakeUpdatePigletDeathUseCase struct {
	death   *pigletdeathdomain.PigletDeath
	err     error
	id      uuid.UUID
	command pigletdeathapplication.UpdatePigletDeathCommand
	called  bool
}

func (f *fakeUpdatePigletDeathUseCase) Execute(_ context.Context, id uuid.UUID, command pigletdeathapplication.UpdatePigletDeathCommand) (*pigletdeathdomain.PigletDeath, error) {
	f.called = true
	f.id = id
	f.command = command
	return f.death, f.err
}

func newTestHandler(create pigletdeathinfra.CreatePigletDeathUseCase, list pigletdeathinfra.ListPigletDeathsUseCase) *pigletdeathinfra.PigletDeathHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return pigletdeathinfra.NewPigletDeathHandler(
		create, list, &fakeDeletePigletDeathUseCase{}, &fakeUpdatePigletDeathUseCase{}, passThrough, passThrough,
	)
}

func newAdminGuardHandler(create pigletdeathinfra.CreatePigletDeathUseCase) *pigletdeathinfra.PigletDeathHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	adminGuard := func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
	}
	return pigletdeathinfra.NewPigletDeathHandler(
		create, &fakeListPigletDeathsUseCase{}, &fakeDeletePigletDeathUseCase{},
		&fakeUpdatePigletDeathUseCase{}, passThrough, adminGuard,
	)
}

func newDeleteHandler(remove pigletdeathinfra.DeletePigletDeathUseCase) *pigletdeathinfra.PigletDeathHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return pigletdeathinfra.NewPigletDeathHandler(
		&fakeCreatePigletDeathUseCase{}, &fakeListPigletDeathsUseCase{}, remove,
		&fakeUpdatePigletDeathUseCase{}, passThrough, passThrough,
	)
}

func newAdminGuardDeleteHandler(remove pigletdeathinfra.DeletePigletDeathUseCase) *pigletdeathinfra.PigletDeathHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	adminGuard := func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
	}
	return pigletdeathinfra.NewPigletDeathHandler(
		&fakeCreatePigletDeathUseCase{}, &fakeListPigletDeathsUseCase{}, remove,
		&fakeUpdatePigletDeathUseCase{}, passThrough, adminGuard,
	)
}

func newUpdateHandler(update pigletdeathinfra.UpdatePigletDeathUseCase) *pigletdeathinfra.PigletDeathHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return pigletdeathinfra.NewPigletDeathHandler(
		&fakeCreatePigletDeathUseCase{}, &fakeListPigletDeathsUseCase{}, &fakeDeletePigletDeathUseCase{},
		update, passThrough, passThrough,
	)
}

func newAdminGuardUpdateHandler(update pigletdeathinfra.UpdatePigletDeathUseCase) *pigletdeathinfra.PigletDeathHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	adminGuard := func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
	}
	return pigletdeathinfra.NewPigletDeathHandler(
		&fakeCreatePigletDeathUseCase{}, &fakeListPigletDeathsUseCase{}, &fakeDeletePigletDeathUseCase{},
		update, passThrough, adminGuard,
	)
}

func authenticatedRequest(request *http.Request, userID uuid.UUID) *http.Request {
	return request.WithContext(authdomain.WithAuthenticatedUser(request.Context(), authdomain.AuthenticatedUser{
		UserID: userID,
		Role:   userdomain.RoleAdmin,
	}))
}

func serve(handler *pigletdeathinfra.PigletDeathHandler, request *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func handlerPigletDeath() *pigletdeathdomain.PigletDeath {
	actor := uuid.New()
	createdAt := time.Date(2026, time.April, 25, 3, 4, 5, 0, time.UTC)
	note := "Muerte durante la noche"
	weight := 2.5
	return &pigletdeathdomain.PigletDeath{
		ID:           uuid.New(),
		FarrowingID:  uuid.New(),
		SowID:        uuid.New(),
		SowCode:      "C-001",
		OperatorID:   uuid.New(),
		OperatorName: "Operador",
		DeathDate:    time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC),
		Quantity:     2,
		Weight:       &weight,
		Cause:        pigletdeathdomain.CauseCrushed,
		Turn:         pigletdeathdomain.TurnMorning,
		Note:         &note,
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
		CreatedBy:    actor,
		UpdatedBy:    actor,
	}
}
