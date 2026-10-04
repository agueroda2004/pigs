package ports

import "errors"

var (
	ErrFarrowingNotFound      = errors.New("Parto no encontrado")
	ErrFarrowingAlreadyExists = errors.New("La cerda ya tiene un parto para este servicio")
	ErrSowNotFound            = errors.New("Cerda no encontrada")
	ErrServiceNotFound        = errors.New("Servicio no encontrado")
	ErrOperatorNotFound       = errors.New("Operador no encontrado")
	ErrMedicationNotFound     = errors.New("Medicamento no encontrado")
)
