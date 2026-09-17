package ports

import "errors"

var (
	ErrRefreshTokenNotFound    = errors.New("Token de refresco no encontrado")
	ErrRefreshTokenAlreadyUsed = errors.New("El token de refresco ya fue utilizado")
)
