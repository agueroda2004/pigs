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
		if body := strings.TrimSpace(response.Body.String()); body != "" {
			t.Fatalf("expected empty body, got %q", body)
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
		if body := strings.TrimSpace(response.Body.String()); body != "" {
			t.Fatalf("expected empty body, got %q", body)
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
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/"+userID.String(), strings.NewReader(`{"name":"New Name","username":"new-user","password":"secret","role":"Admin","active":false}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || update.userID != userID || update.command.UpdatedBy != actorID.String() || update.command.Name == nil || update.command.Username == nil || update.command.Password == nil || update.command.Role == nil || *update.command.Role != role || update.command.Active == nil || *update.command.Active {
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

func TestUserHandlerList(t *testing.T) {
	t.Run("returns every user without passwords", func(t *testing.T) {
		list := &fakeListUsersUseCase{users: []*userdomain.User{handlerUser(), handlerUser()}}
		handler := newTestHandlerWithList(list, &fakeCreateUserUseCase{}, &fakeUpdateUserUseCase{}, &fakeUpdateUserByAdminUseCase{})
		request := authenticatedRequest(httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil), uuid.New())
		response := serve(handler, request)

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		body := response.Body.String()
		if strings.Contains(body, "password") || !strings.HasPrefix(strings.TrimSpace(body), "[") {
			t.Fatalf("unexpected body: %s", body)
		}
		for _, field := range []string{"created_at", "updated_at", "created_by", "updated_by"} {
			if strings.Contains(body, field) {
				t.Fatalf("body should omit audit field %q: %s", field, body)
			}
		}
	})

	t.Run("returns an empty array when there are no users", func(t *testing.T) {
		list := &fakeListUsersUseCase{users: []*userdomain.User{}}
		handler := newTestHandlerWithList(list, &fakeCreateUserUseCase{}, &fakeUpdateUserUseCase{}, &fakeUpdateUserByAdminUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil))

		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		list := &fakeListUsersUseCase{err: errors.New("unexpected")}
		handler := newTestHandlerWithList(list, &fakeCreateUserUseCase{}, &fakeUpdateUserUseCase{}, &fakeUpdateUserByAdminUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}

func TestUserHandlerDelete(t *testing.T) {
	userID := uuid.New()

	t.Run("deletes the user with the provided id", func(t *testing.T) {
		remove := &fakeDeleteUserUseCase{}
		handler := newDeleteHandler(remove)
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/users/"+userID.String(), nil)
		response := serve(handler, request)

		if response.Code != http.StatusNoContent || !remove.called || remove.userID != userID {
			t.Fatalf("status=%d called=%v id=%v", response.Code, remove.called, remove.userID)
		}
		if body := strings.TrimSpace(response.Body.String()); body != "" {
			t.Fatalf("expected empty body, got %q", body)
		}
	})

	t.Run("returns bad request for invalid UUID", func(t *testing.T) {
		remove := &fakeDeleteUserUseCase{}
		handler := newDeleteHandler(remove)
		response := serve(handler, httptest.NewRequest(http.MethodDelete, "/api/v1/admin/users/not-a-uuid", nil))

		if response.Code != http.StatusBadRequest || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("blocks deletion for non-admins", func(t *testing.T) {
		remove := &fakeDeleteUserUseCase{}
		handler := newAdminGuardHandler(remove)
		response := serve(handler, httptest.NewRequest(http.MethodDelete, "/api/v1/admin/users/"+userID.String(), nil))

		if response.Code != http.StatusForbidden || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("maps application errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"not found", ports.ErrUserNotFound, http.StatusNotFound},
			{"in use", ports.ErrUserInUse, http.StatusConflict},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				remove := &fakeDeleteUserUseCase{err: test.err}
				handler := newDeleteHandler(remove)
				response := serve(handler, httptest.NewRequest(http.MethodDelete, "/api/v1/admin/users/"+uuid.New().String(), nil))
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}
