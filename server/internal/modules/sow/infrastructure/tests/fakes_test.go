package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	sowapplication "server/internal/modules/sow/application"
	sowdomain "server/internal/modules/sow/domain"
	sowinfra "server/internal/modules/sow/infrastructure"
	"server/internal/modules/sow/ports"
	userdomain "server/internal/modules/user/domain"
)

type fakeCreateSowUseCase struct {
	sow     *sowdomain.Sow
	err     error
	command sowapplication.CreateSowCommand
	called  bool
}

func (f *fakeCreateSowUseCase) Execute(_ context.Context, command sowapplication.CreateSowCommand) (*sowdomain.Sow, error) {
	f.called = true
	f.command = command
	return f.sow, f.err
}

type fakeListSowsUseCase struct {
	sows   []*sowdomain.Sow
	err    error
	filter ports.SowFilter
	page   int
	called bool
}

func (f *fakeListSowsUseCase) Execute(_ context.Context, filter ports.SowFilter, page int) (sowapplication.SowPage, error) {
	f.called = true
	f.filter = filter
	f.page = page
	return sowapplication.SowPage{
		Items:      f.sows,
		Total:      len(f.sows),
		Page:       page,
		PageSize:   20,
		TotalPages: 1,
	}, f.err
}

type fakeListSowDropdownUseCase struct {
	options []sowdomain.SowDropdown
	err     error
	active  *bool
	states  []sowdomain.State
	called  bool
}

func (f *fakeListSowDropdownUseCase) Execute(_ context.Context, active *bool, states []sowdomain.State) ([]sowdomain.SowDropdown, error) {
	f.called = true
	f.active = active
	f.states = states
	return f.options, f.err
}

type fakeUpdateSowUseCase struct {
	sow     *sowdomain.Sow
	err     error
	sowID   uuid.UUID
	command sowapplication.UpdateSowCommand
	called  bool
}

func (f *fakeUpdateSowUseCase) Execute(_ context.Context, sowID uuid.UUID, command sowapplication.UpdateSowCommand) (*sowdomain.Sow, error) {
	f.called = true
	f.sowID = sowID
	f.command = command
	return f.sow, f.err
}

func newTestHandler(create sowinfra.CreateSowUseCase, list sowinfra.ListSowsUseCase, dropdown sowinfra.ListSowDropdownUseCase, update sowinfra.UpdateSowUseCase) *sowinfra.SowHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return sowinfra.NewSowHandler(create, list, dropdown, update, passThrough, passThrough)
}

func authenticatedRequest(request *http.Request, userID uuid.UUID) *http.Request {
	return request.WithContext(authdomain.WithAuthenticatedUser(request.Context(), authdomain.AuthenticatedUser{
		UserID: userID,
		Role:   userdomain.RoleAdmin,
	}))
}

func serve(handler *sowinfra.SowHandler, request *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func handlerSow() *sowdomain.Sow {
	actor := uuid.New()
	birth := time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC)
	return &sowdomain.Sow{
		ID:        uuid.New(),
		Code:      "C-001",
		Active:    true,
		EntryDate: time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC),
		BirthDate: &birth,
		State:     sowdomain.StateAlive,
		Origin:    sowdomain.OriginOwn,
		Parity:    2,
		BreedID:   uuid.New(),
		CreatedAt: time.Date(2026, time.January, 10, 3, 4, 5, 0, time.UTC),
		UpdatedAt: time.Date(2026, time.January, 10, 3, 4, 5, 0, time.UTC),
		CreatedBy: actor,
		UpdatedBy: actor,
	}
}
