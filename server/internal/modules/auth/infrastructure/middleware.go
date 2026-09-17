package infrastructure

import (
	"net/http"

	authdomain "server/internal/modules/auth/domain"
	"server/internal/modules/auth/ports"
	userdomain "server/internal/modules/user/domain"
)

type Authenticator struct {
	verifier ports.AccessTokenVerifier
}

func NewAuthenticator(verifier ports.AccessTokenVerifier) *Authenticator {
	return &Authenticator{verifier: verifier}
}

func (a *Authenticator) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := readCookie(r, accessTokenCookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, ErrInvalidAccessToken)
			return
		}

		user, err := a.verifier.Verify(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, ErrInvalidAccessToken)
			return
		}

		next.ServeHTTP(w, r.WithContext(authdomain.WithAuthenticatedUser(r.Context(), *user)))
	})
}

func RequireRole(role userdomain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := authdomain.AuthenticatedUserFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusUnauthorized, ErrInvalidAccessToken)
				return
			}
			if user.Role != role {
				writeError(w, http.StatusForbidden, ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAdmin(next http.Handler) http.Handler {
	return RequireRole(userdomain.RoleAdmin)(next)
}
