package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	authapplication "server/internal/modules/auth/application"
	authdomain "server/internal/modules/auth/domain"
	"server/internal/modules/auth/ports"
)

// + === CONSTANTS ===
const maxRequestBodySize = 1 << 20

// + === TYPE ===
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
	accessTTL      time.Duration
	refreshTTL     time.Duration
	cookieConfig   CookieConfig
}

// + === CONSTRUCTOR ===
func NewAuthHandler(
	login LoginUseCase,
	refresh RefreshUseCase,
	logout LogoutUseCase,
	accessTTL time.Duration,
	refreshTTL time.Duration,
	cookieConfig CookieConfig,
) *AuthHandler {
	return &AuthHandler{
		loginUseCase:   login,
		refreshUseCase: refresh,
		logoutUseCase:  logout,
		accessTTL:      accessTTL,
		refreshTTL:     refreshTTL,
		cookieConfig:   cookieConfig,
	}
}

// + === ROUTES ===
func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/login", h.login)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", h.logout)
}

// + === REQUEST ===
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// + === METHODS ===
func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
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

func (h *AuthHandler) refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := readCookie(r, refreshTokenCookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, errors.New("Token de refresco no proporcionado"))
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

func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	if refreshToken, err := readCookie(r, refreshTokenCookieName); err == nil {
		_ = h.logoutUseCase.Execute(r.Context(), authapplication.LogoutCommand{RefreshToken: refreshToken})
	}

	http.SetCookie(w, ClearAccessTokenCookie(h.cookieConfig))
	http.SetCookie(w, ClearRefreshTokenCookie(h.cookieConfig))
	writeOK(w)
}

func (h *AuthHandler) setSessionCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	http.SetCookie(w, BuildAccessTokenCookie(accessToken, int(h.accessTTL.Seconds()), h.cookieConfig))
	http.SetCookie(w, BuildRefreshTokenCookie(refreshToken, int(h.refreshTTL.Seconds()), h.cookieConfig))
}

// + === HELPERS ===
func readCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func writeAuthError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, authapplication.ErrInvalidCredentials),
		errors.Is(err, ports.ErrRefreshTokenNotFound),
		errors.Is(err, authdomain.ErrTokenExpired),
		errors.Is(err, authdomain.ErrTokenRevoked),
		errors.Is(err, authdomain.ErrTokenAlreadyUsed):
		status = http.StatusUnauthorized
	}
	writeError(w, status, err)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return errors.New("El cuerpo de la solicitud no es valido")
	}
	return nil
}

func writeOK(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
