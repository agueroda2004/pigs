package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	userdomain "server/internal/modules/user/domain"
	"server/internal/modules/user/ports"
)

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
