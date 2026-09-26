package ports

import "errors"

var (
	ErrSowRemovalNotFound = errors.New("Baja de cerda no encontrada")
	ErrSowAlreadyRemoved  = errors.New("La cerda ya tiene una baja registrada")
	ErrSowNotFound        = errors.New("Cerda no encontrada")
	ErrServiceNotFound    = errors.New("Servicio no encontrado")
	ErrAbortionNotFound   = errors.New("Aborto no encontrado")
)
