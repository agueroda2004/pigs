package ports

import "errors"

var (
	ErrOperatorNotFound        = errors.New("Operador no encontrado")
	ErrOperatorNameAlreadyUsed = errors.New("El nombre del operador ya existe")
)
