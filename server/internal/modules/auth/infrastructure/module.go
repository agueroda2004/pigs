package infrastructure

import (
	"net/http"
	"time"

	authapplication "server/internal/modules/auth/application"
	"server/internal/modules/auth/ports"
)

type Module struct {
	LoginService    *authapplication.LoginService
	RefreshService  *authapplication.RefreshService
	LogoutService   *authapplication.LogoutService
	Handler         *AuthHandler
	AdminMiddleware func(http.Handler) http.Handler
}

func NewModule(
	users ports.UserReader,
	refreshTokens ports.RefreshTokenRepository,
	passwordVerifier ports.PasswordVerifier,
	accessIssuer ports.AccessTokenIssuer,
	accessVerifier ports.AccessTokenVerifier,
	tokenGenerator ports.TokenGenerator,
	tokenHasher ports.TokenHasher,
	clock func() time.Time,
	accessTTL time.Duration,
	refreshTTL time.Duration,
	cookieConfig CookieConfig,
) *Module {
	login := authapplication.NewLoginService(
		users,
		passwordVerifier,
		accessIssuer,
		tokenGenerator,
		tokenHasher,
		refreshTokens,
		clock,
		refreshTTL,
	)
	refresh := authapplication.NewRefreshService(
		refreshTokens,
		users,
		tokenHasher,
		tokenGenerator,
		accessIssuer,
		clock,
		refreshTTL,
	)
	logout := authapplication.NewLogoutService(refreshTokens, tokenHasher)

	authenticator := NewAuthenticator(accessVerifier)

	return &Module{
		LoginService:   login,
		RefreshService: refresh,
		LogoutService:  logout,
		Handler:        NewAuthHandler(login, refresh, logout, accessTTL, refreshTTL, cookieConfig),
		AdminMiddleware: func(next http.Handler) http.Handler {
			return authenticator.Authenticate(RequireAdmin(next))
		},
	}
}
