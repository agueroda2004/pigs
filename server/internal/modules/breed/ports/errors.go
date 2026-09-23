package ports

import "errors"

var (
	ErrBreedNotFound        = errors.New("Raza no encontrada")
	ErrBreedNameAlreadyUsed = errors.New("El nombre de la raza ya existe")
)
