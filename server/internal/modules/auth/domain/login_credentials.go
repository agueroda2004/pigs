package auth

import (
	"errors"
	"strings"

	userdomain "server/internal/modules/user/domain"
)

var (
	ErrInvalidUsername = errors.New("El nombre de usuario es obligatorio y debe tener como máximo 50 caracteres")
	ErrInvalidPassword = errors.New("La contraseña es obligatoria y debe tener como máximo 50 caracteres")
)

type LoginCredentials struct {
	Username string
	Password string
}

// NewLoginCredentials validates and normalizes the credentials used to log in.
// It trims the username and enforces the shared username and password limits.
func NewLoginCredentials(username string, password string) (LoginCredentials, error) {
	username = strings.TrimSpace(username)
	if username == "" || len([]rune(username)) > userdomain.MaxUsernameLength {
		return LoginCredentials{}, ErrInvalidUsername
	}

	if password == "" || len([]rune(password)) > userdomain.MaxPasswordLength {
		return LoginCredentials{}, ErrInvalidPassword
	}

	return LoginCredentials{Username: username, Password: password}, nil
}
