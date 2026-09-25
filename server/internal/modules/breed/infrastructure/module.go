package infrastructure

import (
	"net/http"
	"time"

	breedapplication "server/internal/modules/breed/application"
	"server/internal/modules/breed/ports"
)

type Module struct {
	CreateBreed      *breedapplication.CreateBreedService
	ListBreeds       *breedapplication.ListBreedsService
	ListBreedOptions *breedapplication.ListBreedOptionsService
	UpdateBreed      *breedapplication.UpdateBreedService
	Handler          *BreedHandler
}

// NewModule assembles the breed use cases and HTTP handler from its dependencies.
// It returns a module exposing the services and handler for wiring.
func NewModule(
	repository ports.BreedRepository,
	clock func() time.Time,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *Module {
	createBreed := breedapplication.NewCreateBreedService(repository, clock)
	listBreeds := breedapplication.NewListBreedsService(repository)
	listBreedOptions := breedapplication.NewListBreedOptionsService(repository)
	updateBreed := breedapplication.NewUpdateBreedService(repository, clock)

	return &Module{
		CreateBreed:      createBreed,
		ListBreeds:       listBreeds,
		ListBreedOptions: listBreedOptions,
		UpdateBreed:      updateBreed,
		Handler:          NewBreedHandler(createBreed, listBreeds, listBreedOptions, updateBreed, authMiddleware, adminMiddleware),
	}
}
