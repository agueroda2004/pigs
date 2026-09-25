package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"

	abortionapplication "server/internal/modules/abortion/application"
	abortiondomain "server/internal/modules/abortion/domain"
	abortioninfra "server/internal/modules/abortion/infrastructure"
	"server/internal/modules/abortion/ports"
	authdomain "server/internal/modules/auth/domain"
	userdomain "server/internal/modules/user/domain"
)

type fakeCreateAbortionUseCase struct {
	abortion *abortiondomain.Abortion
	err      error
	command  abortionapplication.CreateAbortionCommand
	called   bool
}

func (f *fakeCreateAbortionUseCase) Execute(_ context.Context, command abortionapplication.CreateAbortionCommand) (*abortiondomain.Abortion, error) {
	f.called = true
	f.command = command
	return f.abortion, f.err
}

type fakeListAbortionsUseCase struct {
	abortions []*abortiondomain.Abortion
	err       error
	filter    ports.AbortionFilter
	called    bool
}

func (f *fakeListAbortionsUseCase) Execute(_ context.Context, filter ports.AbortionFilter) ([]*abortiondomain.Abortion, error) {
	f.called = true
	f.filter = filter
	return f.abortions, f.err
}

func newTestHandler(create abortioninfra.CreateAbortionUseCase, list abortioninfra.ListAbortionsUseCase) *abortioninfra.AbortionHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return abortioninfra.NewAbortionHandler(create, list, passThrough, passThrough)
}

func authenticatedRequest(request *http.Request, userID uuid.UUID) *http.Request {
	return request.WithContext(authdomain.WithAuthenticatedUser(request.Context(), authdomain.AuthenticatedUser{
		UserID: userID,
		Role:   userdomain.RoleAdmin,
	}))
}

func serve(handler *abortioninfra.AbortionHandler, request *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func handlerAbortion() *abortiondomain.Abortion {
	actor := uuid.New()
	createdAt := time.Date(2026, time.January, 15, 3, 4, 5, 0, time.UTC)
	return &abortiondomain.Abortion{
		ID:           uuid.New(),
		SowID:        uuid.New(),
		ServiceID:    uuid.New(),
		AbortionDate: time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC),
		Cause:        abortiondomain.CauseInfectious,
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
		CreatedBy:    actor,
		UpdatedBy:    actor,
	}
}
