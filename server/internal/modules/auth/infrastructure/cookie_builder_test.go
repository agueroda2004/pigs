package infrastructure

import (
	"net/http"
	"testing"
)

func TestCookieBuilder(t *testing.T) {
	config := CookieConfig{
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	t.Run("builds access token cookie", func(t *testing.T) {
		cookie := BuildAccessTokenCookie("token", 900, config)

		if cookie.Name != "access_token" || cookie.Value != "token" || cookie.Path != "/" || cookie.MaxAge != 900 {
			t.Fatalf("unexpected cookie: %#v", cookie)
		}
		if !cookie.HttpOnly || cookie.Secure || cookie.SameSite != http.SameSiteLaxMode {
			t.Fatalf("unexpected attributes: %#v", cookie)
		}
	})

	t.Run("builds refresh token cookie", func(t *testing.T) {
		cookie := BuildRefreshTokenCookie("token", 604800, config)

		if cookie.Name != "refresh_token" || cookie.Value != "token" || cookie.Path != "/" || cookie.MaxAge != 604800 {
			t.Fatalf("unexpected cookie: %#v", cookie)
		}
		if !cookie.HttpOnly {
			t.Fatalf("refresh cookie must be HttpOnly: %#v", cookie)
		}
	})

	t.Run("builds secure cookies when configured", func(t *testing.T) {
		secureConfig := CookieConfig{Secure: true, SameSite: http.SameSiteStrictMode}

		access := BuildAccessTokenCookie("token", 900, secureConfig)
		refresh := BuildRefreshTokenCookie("token", 604800, secureConfig)

		if !access.Secure || access.SameSite != http.SameSiteStrictMode {
			t.Fatalf("unexpected access cookie: %#v", access)
		}
		if !refresh.Secure {
			t.Fatalf("unexpected refresh cookie: %#v", refresh)
		}
	})

	t.Run("builds clearing cookies", func(t *testing.T) {
		access := ClearAccessTokenCookie(config)
		refresh := ClearRefreshTokenCookie(config)

		if access.Name != "access_token" || access.MaxAge != -1 || access.Value != "" {
			t.Fatalf("unexpected access cookie: %#v", access)
		}
		if refresh.Name != "refresh_token" || refresh.MaxAge != -1 || refresh.Value != "" {
			t.Fatalf("unexpected refresh cookie: %#v", refresh)
		}
	})
}
