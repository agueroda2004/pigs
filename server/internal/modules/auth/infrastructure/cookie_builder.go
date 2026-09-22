package infrastructure

import (
	"net/http"
	"strings"
)

const (
	accessTokenCookieName  = "access_token"
	refreshTokenCookieName = "refresh_token"
)

type CookieConfig struct {
	Secure   bool
	SameSite http.SameSite
}

// SameSiteFromString maps a case-insensitive SameSite value to http.SameSite.
// Unknown or empty values default to http.SameSiteLaxMode.
func SameSiteFromString(value string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

// BuildAccessTokenCookie builds the HttpOnly access token cookie with the given max age.
// SameSite Lax funciona en localhost sobre HTTP. En produccion debe usarse
// Secure=true; si el frontend llegara a ser cross-site, cambiar SameSite a
// http.SameSiteNoneMode junto con Secure (requiere HTTPS).
func BuildAccessTokenCookie(token string, maxAge int, config CookieConfig) *http.Cookie {
	return &http.Cookie{
		Name:     accessTokenCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   config.Secure,
		SameSite: config.SameSite,
		MaxAge:   maxAge,
	}
}

// BuildRefreshTokenCookie builds the HttpOnly refresh token cookie with the given max age.
// It uses the same security attributes as the access token cookie.
func BuildRefreshTokenCookie(token string, maxAge int, config CookieConfig) *http.Cookie {
	return &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   config.Secure,
		SameSite: config.SameSite,
		MaxAge:   maxAge,
	}
}

// ClearAccessTokenCookie builds an expired access token cookie to remove it from the client.
// It sets MaxAge to -1 while keeping the original security attributes.
func ClearAccessTokenCookie(config CookieConfig) *http.Cookie {
	return &http.Cookie{
		Name:     accessTokenCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   config.Secure,
		SameSite: config.SameSite,
		MaxAge:   -1,
	}
}

// ClearRefreshTokenCookie builds an expired refresh token cookie to remove it from the client.
// It sets MaxAge to -1 while keeping the original security attributes.
func ClearRefreshTokenCookie(config CookieConfig) *http.Cookie {
	return &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   config.Secure,
		SameSite: config.SameSite,
		MaxAge:   -1,
	}
}
