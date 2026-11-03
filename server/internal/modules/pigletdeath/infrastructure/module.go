package infrastructure

import (
	"net/http"
	"time"

	pigletdeathapplication "server/internal/modules/pigletdeath/application"
	"server/internal/modules/pigletdeath/ports"
)

type Module struct {
	CreatePigletDeath       *pigletdeathapplication.CreatePigletDeathService
	ListPigletDeaths        *pigletdeathapplication.ListPigletDeathsService
	DeletePigletDeath       *pigletdeathapplication.DeletePigletDeathService
	UpdatePigletDeath       *pigletdeathapplication.UpdatePigletDeathService
	PigletDeathHistoryBySow *pigletdeathapplication.PigletDeathHistoryBySowService
	Handler                 *PigletDeathHandler
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
	deletePigletDeath := pigletdeathapplication.NewDeletePigletDeathService(repository, clock)
	updatePigletDeath := pigletdeathapplication.NewUpdatePigletDeathService(repository, clock)
	pigletDeathHistoryBySow := pigletdeathapplication.NewPigletDeathHistoryBySowService(repository)

	return &Module{
		CreatePigletDeath:       createPigletDeath,
		ListPigletDeaths:        listPigletDeaths,
		DeletePigletDeath:       deletePigletDeath,
		UpdatePigletDeath:       updatePigletDeath,
		PigletDeathHistoryBySow: pigletDeathHistoryBySow,
		Handler: NewPigletDeathHandler(
			createPigletDeath, listPigletDeaths, deletePigletDeath, updatePigletDeath,
			authMiddleware, adminMiddleware,
		),
	}
}
