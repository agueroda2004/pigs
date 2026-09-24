package ports

import "errors"

var (
	ErrSowNotFound        = errors.New("Cerda no encontrada")
	ErrSowCodeAlreadyUsed = errors.New("El código de la cerda ya existe")
)
