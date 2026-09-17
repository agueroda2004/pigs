package ports

import (
	authdomain "server/internal/modules/auth/domain"
)

type AccessTokenIssuer interface {
	Issue(user authdomain.AuthenticatedUser) (string, error)
}

type AccessTokenVerifier interface {
	Verify(token string) (*authdomain.AuthenticatedUser, error)
}
