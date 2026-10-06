package infrastructure

import (
	"net/http"
	"time"

	breedapplication "server/internal/modules/breed/application"
	"server/internal/modules/breed/ports"
)

type Module struct {
	CreateBreed       *breedapplication.CreateBreedService
	ListBreeds        *breedapplication.ListBreedsService
	ListBreedDropdown *breedapplication.ListBreedDropdownService
	UpdateBreed       *breedapplication.UpdateBreedService
	DeleteBreed       *breedapplication.DeleteBreedService
	Handler           *BreedHandler
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
	listBreedDropdown := breedapplication.NewListBreedDropdownService(repository)
	updateBreed := breedapplication.NewUpdateBreedService(repository, clock)
	deleteBreed := breedapplication.NewDeleteBreedService(repository)

	return &Module{
		CreateBreed:       createBreed,
		ListBreeds:        listBreeds,
		ListBreedDropdown: listBreedDropdown,
		UpdateBreed:       updateBreed,
		DeleteBreed:       deleteBreed,
		Handler:           NewBreedHandler(createBreed, listBreeds, listBreedDropdown, updateBreed, deleteBreed, authMiddleware, adminMiddleware),
	}
}
