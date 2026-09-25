package infrastructure

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	authapplication "server/internal/modules/auth/application"
	authdomain "server/internal/modules/auth/domain"
	"server/internal/modules/auth/ports"
	userdomain "server/internal/modules/user/domain"
	userports "server/internal/modules/user/ports"
	platformhttp "server/internal/platform/http"
)

type LoginUseCase interface {
	Execute(context.Context, authapplication.LoginCommand) (*authapplication.LoginResult, error)
}

type RefreshUseCase interface {
	Execute(context.Context, authapplication.RefreshCommand) (*authapplication.RefreshResult, error)
}

type LogoutUseCase interface {
	Execute(context.Context, authapplication.LogoutCommand) error
}

type AuthHandler struct {
	loginUseCase   LoginUseCase
	refreshUseCase RefreshUseCase
	logoutUseCase  LogoutUseCase
	authMiddleware func(http.Handler) http.Handler
	accessTTL      time.Duration
	refreshTTL     time.Duration
	cookieConfig   CookieConfig
}

// NewAuthHandler wires the auth use cases, auth middleware, token TTLs and cookie config.
// It returns a handler ready to register its routes.
func NewAuthHandler(
	login LoginUseCase,
	refresh RefreshUseCase,
	logout LogoutUseCase,
	authMiddleware func(http.Handler) http.Handler,
	accessTTL time.Duration,
	refreshTTL time.Duration,
	cookieConfig CookieConfig,
) *AuthHandler {
	return &AuthHandler{
		loginUseCase:   login,
		refreshUseCase: refresh,
		logoutUseCase:  logout,
		authMiddleware: authMiddleware,
		accessTTL:      accessTTL,
		refreshTTL:     refreshTTL,
		cookieConfig:   cookieConfig,
	}
}

// RegisterRoutes registers the login, refresh, logout and me endpoints on the mux.
// Only the me endpoint is protected by the auth middleware.
func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/login", h.login)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", h.logout)
	mux.Handle("GET /api/v1/auth/me", h.authMiddleware(http.HandlerFunc(h.me)))
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type meResponse struct {
	ID       uuid.UUID       `json:"id"`
	Username string          `json:"username"`
	Role     userdomain.Role `json:"role"`
}

// login handles POST /api/v1/auth/login and starts a session for valid credentials.
// It sets the access and refresh cookies and maps auth errors to HTTP status codes.
func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	result, err := h.loginUseCase.Execute(r.Context(), authapplication.LoginCommand{
		Username: request.Username,
		Password: request.Password,
	})
	if err != nil {
		writeAuthError(w, err)
		return
	}

	h.setSessionCookies(w, result.AccessToken, result.RefreshToken)
	writeOK(w)
}

// refresh handles POST /api/v1/auth/refresh and rotates the session tokens.
// It reads the refresh cookie, executes the use case and sets the new cookies.
func (h *AuthHandler) refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := readCookie(r, refreshTokenCookieName)
	if err != nil {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Token de refresco no proporcionado"))
		return
	}

	result, err := h.refreshUseCase.Execute(r.Context(), authapplication.RefreshCommand{RefreshToken: refreshToken})
	if err != nil {
		writeAuthError(w, err)
		return
	}

	h.setSessionCookies(w, result.AccessToken, result.RefreshToken)
	writeOK(w)
}

// logout handles POST /api/v1/auth/logout and revokes the current token family.
// It always clears the session cookies, even when no refresh token is present.
func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	if refreshToken, err := readCookie(r, refreshTokenCookieName); err == nil {
		_ = h.logoutUseCase.Execute(r.Context(), authapplication.LogoutCommand{RefreshToken: refreshToken})
	}

	http.SetCookie(w, ClearAccessTokenCookie(h.cookieConfig))
	http.SetCookie(w, ClearRefreshTokenCookie(h.cookieConfig))
	writeOK(w)
}

// me handles GET /api/v1/auth/me and returns the authenticated user from the context.
// It relies on the auth middleware, so it responds 401 when no user is present.
func (h *AuthHandler) me(w http.ResponseWriter, r *http.Request) {
	user, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, ErrInvalidAccessToken)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, meResponse{
		ID:       user.UserID,
		Username: user.Username,
		Role:     user.Role,
	})
}

// setSessionCookies writes the access and refresh token cookies using the configured TTLs.
// It converts each TTL to seconds for the cookie max age.
func (h *AuthHandler) setSessionCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	http.SetCookie(w, BuildAccessTokenCookie(accessToken, int(h.accessTTL.Seconds()), h.cookieConfig))
	http.SetCookie(w, BuildRefreshTokenCookie(refreshToken, int(h.refreshTTL.Seconds()), h.cookieConfig))
}

// readCookie returns the value of the named request cookie.
// It returns the lookup error when the cookie is missing.
func readCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// writeAuthError maps auth domain and port errors to HTTP status codes.
// It responds with 401 for credential and token errors and 500 otherwise.
func writeAuthError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, authapplication.ErrInvalidCredentials),
		errors.Is(err, userports.ErrUserInactive),
		errors.Is(err, ports.ErrRefreshTokenNotFound),
		errors.Is(err, authdomain.ErrTokenExpired),
		errors.Is(err, authdomain.ErrTokenRevoked),
		errors.Is(err, authdomain.ErrTokenAlreadyUsed):
		status = http.StatusUnauthorized
	}
	platformhttp.WriteError(w, status, err)
}

// writeOK writes the standard 200 OK JSON response for auth endpoints.
// It returns a simple message body to confirm the operation succeeded.
func writeOK(w http.ResponseWriter) {
	platformhttp.WriteJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}
