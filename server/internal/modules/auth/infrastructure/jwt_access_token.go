package infrastructure

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	userdomain "server/internal/modules/user/domain"
)

var (
	ErrInvalidAccessToken   = errors.New("Token de acceso invalido")
	ErrInvalidSigningMethod = errors.New("Metodo de firma invalido")
	ErrForbidden            = errors.New("No tiene permisos para realizar esta accion")
)

type JWTAccessTokenIssuer struct {
	secret []byte
	ttl    time.Duration
}

// NewJWTAccessTokenIssuer builds a JWT issuer with the given signing secret and token TTL.
// It returns an issuer ready to sign access tokens with HS256.
func NewJWTAccessTokenIssuer(secret string, ttl time.Duration) *JWTAccessTokenIssuer {
	return &JWTAccessTokenIssuer{secret: []byte(secret), ttl: ttl}
}

// Issue signs an HS256 access token with the user id, username, role and expiry claims.
// It returns the signed token string or the signing error.
func (i *JWTAccessTokenIssuer) Issue(user authdomain.AuthenticatedUser) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      user.UserID.String(),
		"username": user.Username,
		"role":     string(user.Role),
		"iat":      now.Unix(),
		"exp":      now.Add(i.ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(i.secret)
}

type JWTAccessTokenVerifier struct {
	secret []byte
}

// NewJWTAccessTokenVerifier builds a JWT verifier with the given signing secret.
// It returns a verifier ready to validate HS256 access tokens.
func NewJWTAccessTokenVerifier(secret string) *JWTAccessTokenVerifier {
	return &JWTAccessTokenVerifier{secret: []byte(secret)}
}

// Verify parses and validates an access token, rejecting any non-HMAC signing method.
// It returns the authenticated user or ErrInvalidAccessToken when the token is invalid.
func (v *JWTAccessTokenVerifier) Verify(token string) (*authdomain.AuthenticatedUser, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return v.secret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, ErrInvalidAccessToken
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidAccessToken
	}

	subject, ok := claims["sub"].(string)
	if !ok {
		return nil, ErrInvalidAccessToken
	}
	userID, err := uuid.Parse(subject)
	if err != nil {
		return nil, ErrInvalidAccessToken
	}

	username, _ := claims["username"].(string)
	role, _ := claims["role"].(string)

	return &authdomain.AuthenticatedUser{
		UserID:   userID,
		Username: strings.TrimSpace(username),
		Role:     userdomain.Role(role),
	}, nil
}
