package ports

import "errors"

var (
	ErrSowNotFound        = errors.New("Cerda no encontrada")
	ErrSowCodeAlreadyUsed = errors.New("El código de la cerda ya existe")
	ErrSowInUse           = errors.New("No se puede eliminar esta cerda porque tiene registros enlazados. Desactívala en su lugar.")
)
