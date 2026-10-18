package ports

import "errors"

var (
	ErrMedicationNotFound        = errors.New("Medicamento no encontrado")
	ErrMedicationNameAlreadyUsed = errors.New("El nombre del medicamento ya existe")
	ErrMedicationInUse           = errors.New("No se puede eliminar este medicamento porque tiene registros enlazados.")
)
