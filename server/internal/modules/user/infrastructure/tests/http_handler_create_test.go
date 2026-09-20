package tests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

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
