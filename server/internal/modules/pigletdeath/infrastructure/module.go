package infrastructure

import (
	"net/http"
	"time"

	pigletdeathapplication "server/internal/modules/pigletdeath/application"
	"server/internal/modules/pigletdeath/ports"
)

type Module struct {
	CreatePigletDeath *pigletdeathapplication.CreatePigletDeathService
	ListPigletDeaths  *pigletdeathapplication.ListPigletDeathsService
	Handler           *PigletDeathHandler
}

// NewModule assembles the piglet death use cases and HTTP handler from its dependencies.
// It returns a module exposing the use cases and handler for wiring.
func NewModule(
	repository ports.PigletDeathRepository,
	clock func() time.Time,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *Module {
	createPigletDeath := pigletdeathapplication.NewCreatePigletDeathService(repository, clock)
	listPigletDeaths := pigletdeathapplication.NewListPigletDeathsService(repository)

	return &Module{
		CreatePigletDeath: createPigletDeath,
		ListPigletDeaths:  listPigletDeaths,
		Handler:           NewPigletDeathHandler(createPigletDeath, listPigletDeaths, authMiddleware, adminMiddleware),
	}
}
