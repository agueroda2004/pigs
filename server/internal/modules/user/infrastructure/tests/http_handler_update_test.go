package tests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"server/internal/modules/user/ports"
)

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
