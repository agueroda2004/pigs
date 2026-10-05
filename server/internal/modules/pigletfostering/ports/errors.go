package ports

import "errors"

var (
	ErrPigletFosteringNotFound = errors.New("Traslado de lechones no encontrado")
	ErrSowNotFound             = errors.New("Cerda no encontrada")
	ErrFarrowingNotFound       = errors.New("Parto no encontrado")
)
