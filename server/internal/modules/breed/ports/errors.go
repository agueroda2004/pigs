package ports

import "errors"

var (
	ErrBreedNotFound        = errors.New("Raza no encontrada")
	ErrBreedNameAlreadyUsed = errors.New("El nombre de la raza ya existe")
	ErrBreedInUse           = errors.New("No se puede eliminar esta raza porque tiene registros enlazados. Desactívala en su lugar.")
)
