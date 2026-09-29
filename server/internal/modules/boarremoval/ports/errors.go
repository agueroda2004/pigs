package ports

import "errors"

var (
	ErrBoarRemovalNotFound = errors.New("Baja de verraco no encontrada")
	ErrBoarAlreadyRemoved  = errors.New("El verraco ya tiene una baja registrada")
	ErrBoarNotFound        = errors.New("Verraco no encontrado")
)
