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

// + === ERRORS ===
var (
	ErrInvalidAccessToken   = errors.New("Token de acceso invalido")
	ErrInvalidSigningMethod = errors.New("Metodo de firma invalido")
	ErrForbidden            = errors.New("No tiene permisos para realizar esta accion")
)

// + === TYPE ===
type JWTAccessTokenIssuer struct {
	secret []byte
	ttl    time.Duration
}

// + === CONSTRUCTOR ===
func NewJWTAccessTokenIssuer(secret string, ttl time.Duration) *JWTAccessTokenIssuer {
	return &JWTAccessTokenIssuer{secret: []byte(secret), ttl: ttl}
}

// + === METHODS ===
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

// ? ==================================================================================================================

// + === TYPE ===
type JWTAccessTokenVerifier struct {
	secret []byte
}

// + === CONSTRUCTOR ===
func NewJWTAccessTokenVerifier(secret string) *JWTAccessTokenVerifier {
	return &JWTAccessTokenVerifier{secret: []byte(secret)}
}

// + === METHODS ===
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
