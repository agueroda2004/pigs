package ports

import "errors"

var (
	ErrBoarNotFound        = errors.New("Verraco no encontrado")
	ErrBoarCodeAlreadyUsed = errors.New("El código del verraco ya existe")
)
