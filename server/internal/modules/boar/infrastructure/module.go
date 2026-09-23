package infrastructure

import (
	"net/http"
	"time"

	boarapplication "server/internal/modules/boar/application"
	"server/internal/modules/boar/ports"
)

type Module struct {
	CreateBoar      *boarapplication.CreateBoarService
	ListBoars       *boarapplication.ListBoarsService
	UpdateBoar      *boarapplication.UpdateBoarService
	ChangeBoarState *boarapplication.ChangeBoarStateService
	Handler         *BoarHandler
}

// NewModule assembles the boar use cases and HTTP handler from its dependencies.
// It returns a module exposing the services and handler for wiring; state changes
// stay internal and are not registered as HTTP routes.
func NewModule(
	repository ports.BoarRepository,
	clock func() time.Time,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *Module {
	createBoar := boarapplication.NewCreateBoarService(repository, clock)
	listBoars := boarapplication.NewListBoarsService(repository)
	updateBoar := boarapplication.NewUpdateBoarService(repository, clock)
	changeBoarState := boarapplication.NewChangeBoarStateService(repository, clock)

	return &Module{
		CreateBoar:      createBoar,
		ListBoars:       listBoars,
		UpdateBoar:      updateBoar,
		ChangeBoarState: changeBoarState,
		Handler:         NewBoarHandler(createBoar, listBoars, updateBoar, authMiddleware, adminMiddleware),
	}
}
