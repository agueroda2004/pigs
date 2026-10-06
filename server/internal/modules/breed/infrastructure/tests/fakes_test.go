package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	breedapplication "server/internal/modules/breed/application"
	breeddomain "server/internal/modules/breed/domain"
	breedinfra "server/internal/modules/breed/infrastructure"
	"server/internal/modules/breed/ports"
	userdomain "server/internal/modules/user/domain"
)

type fakeCreateBreedUseCase struct {
	breed   *breeddomain.Breed
	err     error
	command breedapplication.CreateBreedCommand
	called  bool
}

func (f *fakeCreateBreedUseCase) Execute(_ context.Context, command breedapplication.CreateBreedCommand) (*breeddomain.Breed, error) {
	f.called = true
	f.command = command
	return f.breed, f.err
}

type fakeListBreedsUseCase struct {
	breeds []*breeddomain.Breed
	err    error
	filter ports.BreedFilter
	called bool
}

func (f *fakeListBreedsUseCase) Execute(_ context.Context, filter ports.BreedFilter) ([]*breeddomain.Breed, error) {
	f.called = true
	f.filter = filter
	return f.breeds, f.err
}

type fakeListBreedDropdownUseCase struct {
	options []breeddomain.BreedDropdown
	err     error
	active  bool
	called  bool
}

func (f *fakeListBreedDropdownUseCase) Execute(_ context.Context, active bool) ([]breeddomain.BreedDropdown, error) {
	f.called = true
	f.active = active
	return f.options, f.err
}

type fakeUpdateBreedUseCase struct {
	breed   *breeddomain.Breed
	err     error
	breedID uuid.UUID
	command breedapplication.UpdateBreedCommand
	called  bool
}

func (f *fakeUpdateBreedUseCase) Execute(_ context.Context, breedID uuid.UUID, command breedapplication.UpdateBreedCommand) (*breeddomain.Breed, error) {
	f.called = true
	f.breedID = breedID
	f.command = command
	return f.breed, f.err
}

type fakeDeleteBreedUseCase struct {
	err     error
	breedID uuid.UUID
	called  bool
}

func (f *fakeDeleteBreedUseCase) Execute(_ context.Context, breedID uuid.UUID) error {
	f.called = true
	f.breedID = breedID
	return f.err
}

func newTestHandler(create breedinfra.CreateBreedUseCase, list breedinfra.ListBreedsUseCase, dropdown breedinfra.ListBreedDropdownUseCase, update breedinfra.UpdateBreedUseCase, remove breedinfra.DeleteBreedUseCase) *breedinfra.BreedHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return breedinfra.NewBreedHandler(create, list, dropdown, update, remove, passThrough, passThrough)
}

func newAdminGuardHandler(remove breedinfra.DeleteBreedUseCase) *breedinfra.BreedHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	adminGuard := func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
	}
	return breedinfra.NewBreedHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, &fakeListBreedDropdownUseCase{}, &fakeUpdateBreedUseCase{}, remove, passThrough, adminGuard)
}

func authenticatedRequest(request *http.Request, userID uuid.UUID) *http.Request {
	return request.WithContext(authdomain.WithAuthenticatedUser(request.Context(), authdomain.AuthenticatedUser{
		UserID: userID,
		Role:   userdomain.RoleAdmin,
	}))
}

func serve(handler *breedinfra.BreedHandler, request *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func handlerBreed() *breeddomain.Breed {
	actor := uuid.New()
	return &breeddomain.Breed{
		ID: uuid.New(), Name: "Duroc", Active: true,
		CreatedAt: time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC),
		UpdatedAt: time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC),
		CreatedBy: actor, UpdatedBy: actor,
	}
}
