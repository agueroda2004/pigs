package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	authapplication "server/internal/modules/auth/application"
	authdomain "server/internal/modules/auth/domain"
	authinfra "server/internal/modules/auth/infrastructure"
)

type fakeLoginUseCase struct {
	result  *authapplication.LoginResult
	err     error
	command authapplication.LoginCommand
	called  bool
}

func (f *fakeLoginUseCase) Execute(_ context.Context, command authapplication.LoginCommand) (*authapplication.LoginResult, error) {
	f.called = true
	f.command = command
	return f.result, f.err
}

type fakeRefreshUseCase struct {
	result  *authapplication.RefreshResult
	err     error
	command authapplication.RefreshCommand
	called  bool
}

func (f *fakeRefreshUseCase) Execute(_ context.Context, command authapplication.RefreshCommand) (*authapplication.RefreshResult, error) {
	f.called = true
	f.command = command
	return f.result, f.err
}

type fakeLogoutUseCase struct {
	err     error
	command authapplication.LogoutCommand
	called  bool
}

func (f *fakeLogoutUseCase) Execute(_ context.Context, command authapplication.LogoutCommand) error {
	f.called = true
	f.command = command
	return f.err
}

type fakeAccessTokenVerifier struct {
	user *authdomain.AuthenticatedUser
	err  error
}

func (f *fakeAccessTokenVerifier) Verify(_ string) (*authdomain.AuthenticatedUser, error) {
	return f.user, f.err
}

func newTestAuthHandler(login authinfra.LoginUseCase, refresh authinfra.RefreshUseCase, logout authinfra.LogoutUseCase) *authinfra.AuthHandler {
	return authinfra.NewAuthHandler(
		login,
		refresh,
		logout,
		func(next http.Handler) http.Handler { return next },
		15*time.Minute,
		7*24*time.Hour,
		authinfra.CookieConfig{Secure: false, SameSite: http.SameSiteLaxMode},
	)
}

func serveAuth(handler *authinfra.AuthHandler, request *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}
