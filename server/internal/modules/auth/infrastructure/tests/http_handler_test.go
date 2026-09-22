package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authapplication "server/internal/modules/auth/application"
	authdomain "server/internal/modules/auth/domain"
	"server/internal/modules/auth/ports"
)

func TestAuthHandlerLogin(t *testing.T) {

	t.Run("logs in and sets session cookies", func(t *testing.T) {
		login := &fakeLoginUseCase{
			result: &authapplication.LoginResult{AccessToken: "access", RefreshToken: "refresh"},
		}
		handler := newTestAuthHandler(login, &fakeRefreshUseCase{}, &fakeLogoutUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"ana","password":"secret"}`))
		response := serveAuth(handler, request)

		if response.Code != http.StatusOK || !login.called || login.command.Username != "ana" || login.command.Password != "secret" {
			t.Fatalf("status=%d called=%v command=%#v", response.Code, login.called, login.command)
		}

		cookies := response.Result().Cookies()
		access := findCookie(cookies, "access_token")
		refresh := findCookie(cookies, "refresh_token")
		if access == nil || refresh == nil {
			t.Fatalf("expected session cookies, got %#v", cookies)
		}
		if access.Value != "access" || refresh.Value != "refresh" {
			t.Fatalf("unexpected cookie values: %#v", cookies)
		}
		if !access.HttpOnly || !refresh.HttpOnly {
			t.Fatalf("session cookies must be HttpOnly: %#v", cookies)
		}
		if access.MaxAge != int((15*time.Minute).Seconds()) || refresh.MaxAge != int((7*24*time.Hour).Seconds()) {
			t.Fatalf("unexpected cookie max ages: %#v", cookies)
		}
	})

	t.Run("returns bad request for invalid JSON", func(t *testing.T) {
		login := &fakeLoginUseCase{}
		handler := newTestAuthHandler(login, &fakeRefreshUseCase{}, &fakeLogoutUseCase{})
		response := serveAuth(handler, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":`)))

		if response.Code != http.StatusBadRequest || login.called {
			t.Fatalf("status=%d called=%v", response.Code, login.called)
		}
	})

	t.Run("returns unauthorized for invalid credentials", func(t *testing.T) {
		login := &fakeLoginUseCase{err: authapplication.ErrInvalidCredentials}
		handler := newTestAuthHandler(login, &fakeRefreshUseCase{}, &fakeLogoutUseCase{})
		response := serveAuth(handler, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"ana","password":"wrong"}`)))

		if response.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusUnauthorized)
		}
	})
}

func TestAuthHandlerRefresh(t *testing.T) {
	t.Run("rotates and sets new cookies", func(t *testing.T) {
		refresh := &fakeRefreshUseCase{
			result: &authapplication.RefreshResult{AccessToken: "new-access", RefreshToken: "new-refresh"},
		}
		handler := newTestAuthHandler(&fakeLoginUseCase{}, refresh, &fakeLogoutUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		request.AddCookie(&http.Cookie{Name: "refresh_token", Value: "old-refresh"})
		response := serveAuth(handler, request)

		if response.Code != http.StatusOK || !refresh.called || refresh.command.RefreshToken != "old-refresh" {
			t.Fatalf("status=%d called=%v command=%#v", response.Code, refresh.called, refresh.command)
		}

		cookies := response.Result().Cookies()
		if findCookie(cookies, "access_token") == nil || findCookie(cookies, "refresh_token") == nil {
			t.Fatalf("expected session cookies, got %#v", cookies)
		}
	})

	t.Run("returns unauthorized when cookie is missing", func(t *testing.T) {
		refresh := &fakeRefreshUseCase{}
		handler := newTestAuthHandler(&fakeLoginUseCase{}, refresh, &fakeLogoutUseCase{})
		response := serveAuth(handler, httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil))

		if response.Code != http.StatusUnauthorized || refresh.called {
			t.Fatalf("status=%d called=%v", response.Code, refresh.called)
		}
	})

	t.Run("maps token errors to unauthorized", func(t *testing.T) {
		for _, tt := range []struct {
			name string
			err  error
		}{
			{"not found", ports.ErrRefreshTokenNotFound},
			{"expired", authdomain.ErrTokenExpired},
			{"revoked", authdomain.ErrTokenRevoked},
			{"already used", authdomain.ErrTokenAlreadyUsed},
		} {
			t.Run(tt.name, func(t *testing.T) {
				refresh := &fakeRefreshUseCase{err: tt.err}
				handler := newTestAuthHandler(&fakeLoginUseCase{}, refresh, &fakeLogoutUseCase{})
				request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
				request.AddCookie(&http.Cookie{Name: "refresh_token", Value: "token"})
				response := serveAuth(handler, request)

				if response.Code != http.StatusUnauthorized {
					t.Fatalf("status=%d, want %d", response.Code, http.StatusUnauthorized)
				}
			})
		}
	})
}

func TestAuthHandlerLogout(t *testing.T) {
	t.Run("revokes family and clears cookies", func(t *testing.T) {
		logout := &fakeLogoutUseCase{}
		handler := newTestAuthHandler(&fakeLoginUseCase{}, &fakeRefreshUseCase{}, logout)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		request.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh"})
		response := serveAuth(handler, request)

		if response.Code != http.StatusOK || !logout.called || logout.command.RefreshToken != "refresh" {
			t.Fatalf("status=%d called=%v command=%#v", response.Code, logout.called, logout.command)
		}

		cookies := response.Result().Cookies()
		access := findCookie(cookies, "access_token")
		refresh := findCookie(cookies, "refresh_token")
		if access == nil || refresh == nil || access.MaxAge != -1 || refresh.MaxAge != -1 {
			t.Fatalf("expected clearing cookies, got %#v", cookies)
		}
	})

	t.Run("clears cookies even without a refresh token", func(t *testing.T) {
		logout := &fakeLogoutUseCase{}
		handler := newTestAuthHandler(&fakeLoginUseCase{}, &fakeRefreshUseCase{}, logout)
		response := serveAuth(handler, httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil))

		if response.Code != http.StatusOK || logout.called {
			t.Fatalf("status=%d called=%v", response.Code, logout.called)
		}
	})
}
