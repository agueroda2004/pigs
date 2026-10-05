package infrastructure

import (
	"net/http"
	"time"

	pigletfosteringapplication "server/internal/modules/pigletfostering/application"
	"server/internal/modules/pigletfostering/ports"
)

type Module struct {
	CreatePigletFostering *pigletfosteringapplication.CreatePigletFosteringService
	ListPigletFosterings  *pigletfosteringapplication.ListPigletFosteringsService
	Handler               *PigletFosteringHandler
}

// NewModule assembles the fostering use cases and HTTP handler from its dependencies.
// It returns a module exposing the use cases and handler for wiring.
func NewModule(
	repository ports.PigletFosteringRepository,
	clock func() time.Time,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *Module {
	createPigletFostering := pigletfosteringapplication.NewCreatePigletFosteringService(repository, clock)
	listPigletFosterings := pigletfosteringapplication.NewListPigletFosteringsService(repository)

	return &Module{
		CreatePigletFostering: createPigletFostering,
		ListPigletFosterings:  listPigletFosterings,
		Handler:               NewPigletFosteringHandler(createPigletFostering, listPigletFosterings, authMiddleware, adminMiddleware),
	}
}
