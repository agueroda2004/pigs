package ports

import "errors"

var (
	ErrAbortionNotFound = errors.New("Aborto no encontrado")
	ErrSowNotFound      = errors.New("Cerda no encontrada")
	ErrServiceNotFound  = errors.New("Servicio no encontrado")
	// ErrAbortionNotDeletable is returned when the sow already has events after
	// the abortion (a later mount or a sow removal), so it cannot be deleted.
	ErrAbortionNotDeletable = errors.New("El aborto tiene registros posteriores y no se puede eliminar")
)
