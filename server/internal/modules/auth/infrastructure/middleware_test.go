package infrastructure

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	userdomain "server/internal/modules/user/domain"
)

type fakeAccessTokenVerifier struct {
	user *authdomain.AuthenticatedUser
	err  error
}

func (f *fakeAccessTokenVerifier) Verify(_ string) (*authdomain.AuthenticatedUser, error) {
	return f.user, f.err
}

func TestAuthenticatorAuthenticate(t *testing.T) {
	user := &authdomain.AuthenticatedUser{UserID: uuid.New(), Username: "ana", Role: userdomain.RoleAdmin}
	authenticator := NewAuthenticator(&fakeAccessTokenVerifier{user: user})

	t.Run("passes and stores user in context", func(t *testing.T) {
		var got authdomain.AuthenticatedUser
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got, _ = authdomain.AuthenticatedUserFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.AddCookie(&http.Cookie{Name: accessTokenCookieName, Value: "token"})
		response := httptest.NewRecorder()
		authenticator.Authenticate(next).ServeHTTP(response, request)

		if response.Code != http.StatusOK || got.UserID != user.UserID || got.Role != user.Role {
			t.Fatalf("status=%d user=%#v", response.Code, got)
		}
	})

	t.Run("returns unauthorized without cookie", func(t *testing.T) {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()
		authenticator.Authenticate(next).ServeHTTP(response, request)

		if response.Code != http.StatusUnauthorized || called {
			t.Fatalf("status=%d called=%v", response.Code, called)
		}
	})

	t.Run("returns unauthorized for invalid token", func(t *testing.T) {
		bad := NewAuthenticator(&fakeAccessTokenVerifier{err: ErrInvalidAccessToken})
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.AddCookie(&http.Cookie{Name: accessTokenCookieName, Value: "bad"})
		response := httptest.NewRecorder()
		bad.Authenticate(next).ServeHTTP(response, request)

		if response.Code != http.StatusUnauthorized || called {
			t.Fatalf("status=%d called=%v", response.Code, called)
		}
	})
}

func TestRequireRole(t *testing.T) {
	admin := authdomain.AuthenticatedUser{UserID: uuid.New(), Role: userdomain.RoleAdmin}
	regular := authdomain.AuthenticatedUser{UserID: uuid.New(), Role: userdomain.RoleUser}

	t.Run("passes for matching role", func(t *testing.T) {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true; w.WriteHeader(http.StatusOK) })

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request = request.WithContext(authdomain.WithAuthenticatedUser(request.Context(), admin))
		response := httptest.NewRecorder()
		RequireAdmin(next).ServeHTTP(response, request)

		if response.Code != http.StatusOK || !called {
			t.Fatalf("status=%d called=%v", response.Code, called)
		}
	})

	t.Run("returns forbidden for wrong role", func(t *testing.T) {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request = request.WithContext(authdomain.WithAuthenticatedUser(request.Context(), regular))
		response := httptest.NewRecorder()
		RequireAdmin(next).ServeHTTP(response, request)

		if response.Code != http.StatusForbidden || called {
			t.Fatalf("status=%d called=%v", response.Code, called)
		}
	})

	t.Run("returns unauthorized without user in context", func(t *testing.T) {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()
		RequireAdmin(next).ServeHTTP(response, request)

		if response.Code != http.StatusUnauthorized || called {
			t.Fatalf("status=%d called=%v", response.Code, called)
		}
	})
}
