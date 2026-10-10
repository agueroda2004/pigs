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
	total     int
	err       error
	filter    ports.AbortionFilter
	page      int
	called    bool
}

type fakeUpdateAbortionUseCase struct {
	err     error
	id      uuid.UUID
	command abortionapplication.UpdateAbortionCommand
	called  bool
}

func (f *fakeUpdateAbortionUseCase) Execute(_ context.Context, id uuid.UUID, command abortionapplication.UpdateAbortionCommand) error {
	f.called = true
	f.id = id
	f.command = command
	return f.err
}

type fakeDeleteAbortionUseCase struct {
	err     error
	command abortionapplication.DeleteAbortionCommand
	called  bool
}

func (f *fakeDeleteAbortionUseCase) Execute(_ context.Context, command abortionapplication.DeleteAbortionCommand) error {
	f.called = true
	f.command = command
	return f.err
}

func (f *fakeListAbortionsUseCase) Execute(_ context.Context, filter ports.AbortionFilter, page int) (abortionapplication.AbortionPage, error) {
	f.called = true
	f.filter = filter
	f.page = page
	total := f.total
	if total == 0 && len(f.abortions) > 0 {
		total = len(f.abortions)
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + abortionapplication.DefaultAbortionPageSize - 1) / abortionapplication.DefaultAbortionPageSize
	}
	return abortionapplication.AbortionPage{
		Items:      f.abortions,
		Total:      total,
		Page:       page,
		PageSize:   abortionapplication.DefaultAbortionPageSize,
		TotalPages: totalPages,
	}, f.err
}

func newTestHandler(create abortioninfra.CreateAbortionUseCase, list abortioninfra.ListAbortionsUseCase) *abortioninfra.AbortionHandler {
	return newTestHandlerWithDeps(create, list, &fakeUpdateAbortionUseCase{}, &fakeDeleteAbortionUseCase{})
}

func newTestHandlerWithUpdate(create abortioninfra.CreateAbortionUseCase, list abortioninfra.ListAbortionsUseCase, update abortioninfra.UpdateAbortionUseCase) *abortioninfra.AbortionHandler {
	return newTestHandlerWithDeps(create, list, update, &fakeDeleteAbortionUseCase{})
}

func newTestHandlerWithDelete(create abortioninfra.CreateAbortionUseCase, list abortioninfra.ListAbortionsUseCase, remove abortioninfra.DeleteAbortionUseCase) *abortioninfra.AbortionHandler {
	return newTestHandlerWithDeps(create, list, &fakeUpdateAbortionUseCase{}, remove)
}

func newTestHandlerWithDeps(create abortioninfra.CreateAbortionUseCase, list abortioninfra.ListAbortionsUseCase, update abortioninfra.UpdateAbortionUseCase, remove abortioninfra.DeleteAbortionUseCase) *abortioninfra.AbortionHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return abortioninfra.NewAbortionHandler(create, list, update, remove, passThrough, passThrough)
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
		SowCode:      "C-001",
		ServiceID:    uuid.New(),
		AbortionDate: time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC),
		Cause:        abortiondomain.CauseInfectious,
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
		CreatedBy:    actor,
		UpdatedBy:    actor,
	}
}
