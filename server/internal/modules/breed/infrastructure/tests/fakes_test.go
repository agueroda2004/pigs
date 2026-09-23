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
	called bool
}

func (f *fakeListBreedsUseCase) Execute(_ context.Context) ([]*breeddomain.Breed, error) {
	f.called = true
	return f.breeds, f.err
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

func newTestHandler(create breedinfra.CreateBreedUseCase, list breedinfra.ListBreedsUseCase, update breedinfra.UpdateBreedUseCase) *breedinfra.BreedHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return breedinfra.NewBreedHandler(create, list, update, passThrough, passThrough)
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
