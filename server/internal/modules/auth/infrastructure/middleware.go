package infrastructure

import (
	"net/http"

	authdomain "server/internal/modules/auth/domain"
	"server/internal/modules/auth/ports"
	userdomain "server/internal/modules/user/domain"
	platformhttp "server/internal/platform/http"
)

type Authenticator struct {
	verifier ports.AccessTokenVerifier
}

// NewAuthenticator builds the middleware that authenticates requests by access token.
// It returns an authenticator backed by the given token verifier.
func NewAuthenticator(verifier ports.AccessTokenVerifier) *Authenticator {
	return &Authenticator{verifier: verifier}
}

// Authenticate reads the access token cookie, verifies it and stores the user in the context.
// It responds with 401 when the cookie is missing or the token is invalid.
func (a *Authenticator) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := readCookie(r, accessTokenCookieName)
		if err != nil {
			platformhttp.WriteError(w, http.StatusUnauthorized, ErrInvalidAccessToken)
			return
		}

		user, err := a.verifier.Verify(token)
		if err != nil {
			platformhttp.WriteError(w, http.StatusUnauthorized, ErrInvalidAccessToken)
			return
		}

		next.ServeHTTP(w, r.WithContext(authdomain.WithAuthenticatedUser(r.Context(), *user)))
	})
}

// RequireRole builds a middleware that allows only users holding the given role.
// It responds with 401 when no user is in the context and 403 for a wrong role.
func RequireRole(role userdomain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := authdomain.AuthenticatedUserFromContext(r.Context())
			if !ok {
				platformhttp.WriteError(w, http.StatusUnauthorized, ErrInvalidAccessToken)
				return
			}
			if user.Role != role {
				platformhttp.WriteError(w, http.StatusForbidden, ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin builds a middleware that allows only users holding the admin role.
// It is a convenience wrapper around RequireRole for userdomain.RoleAdmin.
func RequireAdmin(next http.Handler) http.Handler {
	return RequireRole(userdomain.RoleAdmin)(next)
}
