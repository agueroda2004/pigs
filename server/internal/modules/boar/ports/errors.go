package ports

import "errors"

var (
	ErrBoarNotFound        = errors.New("Verraco no encontrado")
	ErrBoarCodeAlreadyUsed = errors.New("El código del verraco ya existe")
	ErrBoarInUse           = errors.New("No se puede eliminar este verraco porque tiene registros enlazados. Desactívalo en su lugar.")
)
