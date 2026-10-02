package ports

import "errors"

var (
	ErrMedicationNotFound        = errors.New("Medicamento no encontrado")
	ErrMedicationNameAlreadyUsed = errors.New("El nombre del medicamento ya existe")
)
