package ports

import "errors"

var (
	ErrFarrowingNotFound      = errors.New("Parto no encontrado")
	ErrFarrowingAlreadyExists = errors.New("La cerda ya tiene un parto para este servicio")
	ErrSowNotFound            = errors.New("Cerda no encontrada")
	ErrServiceNotFound        = errors.New("Servicio no encontrado")
	ErrOperatorNotFound       = errors.New("Operador no encontrado")
	ErrMedicationNotFound     = errors.New("Medicamento no encontrado")
	// ErrFarrowingHasRelatedRecords is returned when a farrowing cannot be deleted
	// because it still has deaths, fosterings or weagings linked to it.
	ErrFarrowingHasRelatedRecords = errors.New("Este parto tiene registros enlazados (muertes, adopciones o destetes de lechones); elimínalos antes de borrar el parto")
)
