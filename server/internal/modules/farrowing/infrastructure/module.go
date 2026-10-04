package infrastructure

import (
	"net/http"
	"time"

	farrowingapplication "server/internal/modules/farrowing/application"
	"server/internal/modules/farrowing/ports"
)

type Module struct {
	CreateFarrowing *farrowingapplication.CreateFarrowingService
	ListFarrowings  *farrowingapplication.ListFarrowingsService
	Handler         *FarrowingHandler
}

// NewModule assembles the farrowing use cases and HTTP handler from its dependencies.
// It returns a module exposing the use cases and handler for wiring.
func NewModule(
	repository ports.FarrowingRepository,
	clock func() time.Time,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *Module {
	createFarrowing := farrowingapplication.NewCreateFarrowingService(repository, clock)
	listFarrowings := farrowingapplication.NewListFarrowingsService(repository)

	return &Module{
		CreateFarrowing: createFarrowing,
		ListFarrowings:  listFarrowings,
		Handler:         NewFarrowingHandler(createFarrowing, listFarrowings, authMiddleware, adminMiddleware),
	}
}
