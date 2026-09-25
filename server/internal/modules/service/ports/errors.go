package ports

import "errors"

var (
	ErrServiceNotFound  = errors.New("Servicio no encontrado")
	ErrSowNotFound      = errors.New("Cerda no encontrada")
	ErrBoarNotFound     = errors.New("Verraco no encontrado")
	ErrOperatorNotFound = errors.New("Operador no encontrado")
)
