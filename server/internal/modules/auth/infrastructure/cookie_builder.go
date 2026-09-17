package infrastructure

import (
	"net/http"
	"strings"
)

// + === CONSTANTS ===
const (
	accessTokenCookieName  = "access_token"
	refreshTokenCookieName = "refresh_token"
)

// + === TYPE ===
type CookieConfig struct {
	Secure   bool
	SameSite http.SameSite
}

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

// + === CONSTRUCTOR ===
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
