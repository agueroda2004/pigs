package ports

import "errors"

var (
	ErrUserNotFound        = errors.New("Usuario no encontrado")
	ErrUsernameAlreadyUsed = errors.New("El nombre de usuario ya existe")
	ErrUserInactive        = errors.New("El usuario está inactivo")
	ErrUserInUse           = errors.New("No se puede eliminar este usuario porque tiene registros enlazados.")
)
