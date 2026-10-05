package ports

import "errors"

var (
	ErrPigletDeathNotFound = errors.New("Muerte de lechones no encontrada")
	ErrSowNotFound         = errors.New("Cerda no encontrada")
	ErrFarrowingNotFound   = errors.New("Parto no encontrado")
	ErrOperatorNotFound    = errors.New("Operador no encontrado")
)
