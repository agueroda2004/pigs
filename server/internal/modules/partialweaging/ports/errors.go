package ports

import "errors"

var (
	ErrPartialWeagingNotFound = errors.New("Destete parcial no encontrado")
	ErrSowNotFound            = errors.New("Cerda no encontrada")
	ErrFarrowingNotFound      = errors.New("Parto no encontrado")
)
