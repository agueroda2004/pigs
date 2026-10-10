package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	farrowingapplication "server/internal/modules/farrowing/application"
	farrowingdomain "server/internal/modules/farrowing/domain"
	farrowinginfra "server/internal/modules/farrowing/infrastructure"
	"server/internal/modules/farrowing/ports"
	userdomain "server/internal/modules/user/domain"
)

type fakeCreateFarrowingUseCase struct {
	farrowing *farrowingdomain.Farrowing
	err       error
	command   farrowingapplication.CreateFarrowingCommand
	called    bool
}

func (f *fakeCreateFarrowingUseCase) Execute(_ context.Context, command farrowingapplication.CreateFarrowingCommand) (*farrowingdomain.Farrowing, error) {
	f.called = true
	f.command = command
	return f.farrowing, f.err
}

type fakeListFarrowingsUseCase struct {
	farrowings []*farrowingdomain.Farrowing
	total      int
	err        error
	filter     ports.FarrowingFilter
	page       int
	called     bool
}

func (f *fakeListFarrowingsUseCase) Execute(_ context.Context, filter ports.FarrowingFilter, page int) (farrowingapplication.FarrowingPage, error) {
	f.called = true
	f.filter = filter
	f.page = page
	if f.err != nil {
		return farrowingapplication.FarrowingPage{}, f.err
	}

	total := f.total
	if total == 0 {
		total = len(f.farrowings)
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + farrowingapplication.DefaultFarrowingPageSize - 1) / farrowingapplication.DefaultFarrowingPageSize
	}

	return farrowingapplication.FarrowingPage{
		Items:      f.farrowings,
		Total:      total,
		Page:       page,
		PageSize:   farrowingapplication.DefaultFarrowingPageSize,
		TotalPages: totalPages,
	}, nil
}

type fakeUpdateFarrowingUseCase struct {
	err     error
	id      uuid.UUID
	command farrowingapplication.UpdateFarrowingCommand
	called  bool
}

func (f *fakeUpdateFarrowingUseCase) Execute(_ context.Context, id uuid.UUID, command farrowingapplication.UpdateFarrowingCommand) error {
	f.called = true
	f.id = id
	f.command = command
	return f.err
}

type fakeDeleteFarrowingUseCase struct {
	err     error
	command farrowingapplication.DeleteFarrowingCommand
	called  bool
}

func (f *fakeDeleteFarrowingUseCase) Execute(_ context.Context, command farrowingapplication.DeleteFarrowingCommand) error {
	f.called = true
	f.command = command
	return f.err
}

func newTestHandler(create farrowinginfra.CreateFarrowingUseCase, list farrowinginfra.ListFarrowingsUseCase) *farrowinginfra.FarrowingHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return farrowinginfra.NewFarrowingHandler(create, list, &fakeUpdateFarrowingUseCase{}, &fakeDeleteFarrowingUseCase{}, passThrough, passThrough)
}

func newAdminGuardHandler(create farrowinginfra.CreateFarrowingUseCase) *farrowinginfra.FarrowingHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	adminGuard := func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
	}
	return farrowinginfra.NewFarrowingHandler(create, &fakeListFarrowingsUseCase{}, &fakeUpdateFarrowingUseCase{}, &fakeDeleteFarrowingUseCase{}, passThrough, adminGuard)
}

func newUpdateHandler(update farrowinginfra.UpdateFarrowingUseCase) *farrowinginfra.FarrowingHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return farrowinginfra.NewFarrowingHandler(&fakeCreateFarrowingUseCase{}, &fakeListFarrowingsUseCase{}, update, &fakeDeleteFarrowingUseCase{}, passThrough, passThrough)
}

func newDeleteHandler(remove farrowinginfra.DeleteFarrowingUseCase) *farrowinginfra.FarrowingHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return farrowinginfra.NewFarrowingHandler(&fakeCreateFarrowingUseCase{}, &fakeListFarrowingsUseCase{}, &fakeUpdateFarrowingUseCase{}, remove, passThrough, passThrough)
}

func newAdminGuardUpdateHandler(update farrowinginfra.UpdateFarrowingUseCase) *farrowinginfra.FarrowingHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	adminGuard := func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
	}
	return farrowinginfra.NewFarrowingHandler(&fakeCreateFarrowingUseCase{}, &fakeListFarrowingsUseCase{}, update, &fakeDeleteFarrowingUseCase{}, passThrough, adminGuard)
}

func newAdminGuardDeleteHandler(remove farrowinginfra.DeleteFarrowingUseCase) *farrowinginfra.FarrowingHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	adminGuard := func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
	}
	return farrowinginfra.NewFarrowingHandler(&fakeCreateFarrowingUseCase{}, &fakeListFarrowingsUseCase{}, &fakeUpdateFarrowingUseCase{}, remove, passThrough, adminGuard)
}

func authenticatedRequest(request *http.Request, userID uuid.UUID) *http.Request {
	return request.WithContext(authdomain.WithAuthenticatedUser(request.Context(), authdomain.AuthenticatedUser{
		UserID: userID,
		Role:   userdomain.RoleAdmin,
	}))
}

func serve(handler *farrowinginfra.FarrowingHandler, request *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func handlerFarrowing() *farrowingdomain.Farrowing {
	actor := uuid.New()
	createdAt := time.Date(2026, time.April, 20, 3, 4, 5, 0, time.UTC)
	location := "Corral 3"
	start := "22:00"
	end := "02:00"
	return &farrowingdomain.Farrowing{
		ID:             uuid.New(),
		SowID:          uuid.New(),
		ServiceID:      uuid.New(),
		FarrowDate:     time.Date(2026, time.April, 20, 0, 0, 0, 0, time.UTC),
		StartTime:      &start,
		EndTime:        &end,
		Location:       &location,
		LiveBorn:       10,
		Stillborn:      1,
		Mummified:      0,
		CurrentPiglets: 10,
		IsManipulated:  true,
		Operators:      []*farrowingdomain.FarrowingOperator{},
		Medications:    []*farrowingdomain.FarrowingMedication{},
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
		CreatedBy:      actor,
		UpdatedBy:      actor,
	}
}
