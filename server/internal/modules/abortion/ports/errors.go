package ports

import "errors"

var (
	ErrAbortionNotFound = errors.New("Aborto no encontrado")
	ErrSowNotFound      = errors.New("Cerda no encontrada")
	ErrServiceNotFound  = errors.New("Servicio no encontrado")
)
