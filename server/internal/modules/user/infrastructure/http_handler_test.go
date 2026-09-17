package infrastructure

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	userapplication "server/internal/modules/user/application"
	userdomain "server/internal/modules/user/domain"
	"server/internal/modules/user/ports"
)

func TestUserHandlerCreate(t *testing.T) {
	user := handlerUser()
	actorID := uuid.New()

	t.Run("creates user with actor from context", func(t *testing.T) {
		create := &fakeCreateUserUseCase{user: user}
		handler := newTestHandler(create, &fakeUpdateUserUseCase{}, &fakeUpdateUserByAdminUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(`{"name":"Ana","username":"ana","password":"secret","role":"User"}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusCreated || create.command.CreatedBy != actorID.String() || create.command.Username != "ana" {
			t.Fatalf("status=%d command=%#v", response.Code, create.command)
		}
		if response.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type: %q", response.Header().Get("Content-Type"))
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		create := &fakeCreateUserUseCase{user: user}
		handler := newTestHandler(create, &fakeUpdateUserUseCase{}, &fakeUpdateUserByAdminUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(`{"name":"Ana","username":"ana","password":"secret"}`)))

		if response.Code != http.StatusUnauthorized || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid JSON", func(t *testing.T) {
		create := &fakeCreateUserUseCase{user: user}
		handler := newTestHandler(create, &fakeUpdateUserUseCase{}, &fakeUpdateUserByAdminUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(`{"name":`)))

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("maps application errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"duplicate", ports.ErrUsernameAlreadyUsed, http.StatusConflict},
			{"validation", userdomain.ErrInvalidName, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				create := &fakeCreateUserUseCase{err: test.err}
				handler := newTestHandler(create, &fakeUpdateUserUseCase{}, &fakeUpdateUserByAdminUseCase{})
				request := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(`{"name":"Ana"}`))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestUserHandlerUpdate(t *testing.T) {
	userID := uuid.New()

	t.Run("updates user", func(t *testing.T) {
		update := &fakeUpdateUserUseCase{user: handlerUser()}
		handler := newTestHandler(&fakeCreateUserUseCase{}, update, &fakeUpdateUserByAdminUseCase{})
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+userID.String(), strings.NewReader(`{"name":"New Name"}`))
		response := serve(handler, request)

		if response.Code != http.StatusOK || update.userID != userID || update.command.Name == nil || *update.command.Name != "New Name" {
			t.Fatalf("status=%d id=%v command=%#v", response.Code, update.userID, update.command)
		}
	})

	t.Run("returns bad request for invalid UUID", func(t *testing.T) {
		update := &fakeUpdateUserUseCase{}
		handler := newTestHandler(&fakeCreateUserUseCase{}, update, &fakeUpdateUserByAdminUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPatch, "/api/v1/users/not-a-uuid", strings.NewReader(`{}`)))

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("maps not found and internal errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"not found", ports.ErrUserNotFound, http.StatusNotFound},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				update := &fakeUpdateUserUseCase{err: test.err}
				handler := newTestHandler(&fakeCreateUserUseCase{}, update, &fakeUpdateUserByAdminUseCase{})
				response := serve(handler, httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+uuid.New().String(), strings.NewReader(`{"name":"Ana"}`)))
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestUserHandlerUpdateByAdmin(t *testing.T) {
	userID := uuid.New()
	actorID := uuid.New()
	role := userdomain.RoleAdmin

	t.Run("updates all administrative fields and forwards actor", func(t *testing.T) {
		update := &fakeUpdateUserByAdminUseCase{user: handlerUser()}
		handler := newTestHandler(&fakeCreateUserUseCase{}, &fakeUpdateUserUseCase{}, update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/"+userID.String(), strings.NewReader(`{"name":"New Name","username":"new-user","password":"secret","role":"Admin"}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || update.userID != userID || update.command.UpdatedBy != actorID.String() || update.command.Name == nil || update.command.Username == nil || update.command.Password == nil || update.command.Role == nil || *update.command.Role != role {
			t.Fatalf("status=%d id=%v command=%#v", response.Code, update.userID, update.command)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		update := &fakeUpdateUserByAdminUseCase{user: handlerUser()}
		handler := newTestHandler(&fakeCreateUserUseCase{}, &fakeUpdateUserUseCase{}, update)
		response := serve(handler, httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/"+userID.String(), strings.NewReader(`{"name":"New Name"}`)))

		if response.Code != http.StatusUnauthorized || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("maps invalid UUID and application errors", func(t *testing.T) {
		update := &fakeUpdateUserByAdminUseCase{}
		handler := newTestHandler(&fakeCreateUserUseCase{}, &fakeUpdateUserUseCase{}, update)
		response := serve(handler, httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/not-a-uuid", strings.NewReader(`{"role":"Admin"}`)))
		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("invalid UUID status=%d called=%v", response.Code, update.called)
		}

		update = &fakeUpdateUserByAdminUseCase{err: ports.ErrUsernameAlreadyUsed}
		handler = newTestHandler(&fakeCreateUserUseCase{}, &fakeUpdateUserUseCase{}, update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/"+userID.String(), strings.NewReader(`{"username":"taken"}`))
		request = authenticatedRequest(request, actorID)
		response = serve(handler, request)
		if response.Code != http.StatusConflict {
			t.Fatalf("duplicate status=%d", response.Code)
		}
	})
}

type fakeCreateUserUseCase struct {
	user    *userdomain.User
	err     error
	command userapplication.CreateUserCommand
	called  bool
}

func (f *fakeCreateUserUseCase) Execute(_ context.Context, command userapplication.CreateUserCommand) (*userdomain.User, error) {
	f.called = true
	f.command = command
	return f.user, f.err
}

type fakeUpdateUserUseCase struct {
	user    *userdomain.User
	err     error
	userID  uuid.UUID
	command userapplication.UpdateOwnUserCommand
	called  bool
}

type fakeUpdateUserByAdminUseCase struct {
	user    *userdomain.User
	err     error
	userID  uuid.UUID
	command userapplication.UpdateUserByAdminCommand
	called  bool
}

func (f *fakeUpdateUserUseCase) Execute(_ context.Context, userID uuid.UUID, command userapplication.UpdateOwnUserCommand) (*userdomain.User, error) {
	f.called = true
	f.userID = userID
	f.command = command
	return f.user, f.err
}

func (f *fakeUpdateUserByAdminUseCase) Execute(_ context.Context, userID uuid.UUID, command userapplication.UpdateUserByAdminCommand) (*userdomain.User, error) {
	f.called = true
	f.userID = userID
	f.command = command
	return f.user, f.err
}

func newTestHandler(create CreateUserUseCase, update UpdateOwnUserUseCase, updateAdmin UpdateUserByAdminUseCase) *UserHandler {
	return NewUserHandler(create, update, updateAdmin, func(next http.Handler) http.Handler { return next })
}

func authenticatedRequest(request *http.Request, userID uuid.UUID) *http.Request {
	return request.WithContext(authdomain.WithAuthenticatedUser(request.Context(), authdomain.AuthenticatedUser{
		UserID: userID,
		Role:   userdomain.RoleAdmin,
	}))
}

func serve(handler *UserHandler, request *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func handlerUser() *userdomain.User {
	return &userdomain.User{
		ID: uuid.New(), Name: "Ana", Username: "ana", Role: userdomain.RoleUser,
		CreatedAt: time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC),
		UpdatedAt: time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC),
	}
}
