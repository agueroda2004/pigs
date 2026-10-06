package ports

import "errors"

var (
	ErrWeagingNotFound      = errors.New("Destete no encontrado")
	ErrWeagingAlreadyExists = errors.New("El parto ya tiene un destete registrado")
	ErrSowNotFound          = errors.New("Cerda no encontrada")
	ErrFarrowingNotFound    = errors.New("Parto no encontrado")
)
