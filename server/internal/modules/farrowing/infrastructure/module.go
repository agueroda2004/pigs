package infrastructure

import (
	"net/http"
	"time"

	farrowingapplication "server/internal/modules/farrowing/application"
	"server/internal/modules/farrowing/ports"
)

type Module struct {
	CreateFarrowing       *farrowingapplication.CreateFarrowingService
	ListFarrowings        *farrowingapplication.ListFarrowingsService
	UpdateFarrowing       *farrowingapplication.UpdateFarrowingService
	DeleteFarrowing       *farrowingapplication.DeleteFarrowingService
	FarrowingHistoryBySow *farrowingapplication.FarrowingHistoryBySowService
	Handler               *FarrowingHandler
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
	updateFarrowing := farrowingapplication.NewUpdateFarrowingService(repository, clock)
	deleteFarrowing := farrowingapplication.NewDeleteFarrowingService(repository, clock)
	farrowingHistoryBySow := farrowingapplication.NewFarrowingHistoryBySowService(repository)

	return &Module{
		CreateFarrowing:       createFarrowing,
		ListFarrowings:        listFarrowings,
		UpdateFarrowing:       updateFarrowing,
		DeleteFarrowing:       deleteFarrowing,
		FarrowingHistoryBySow: farrowingHistoryBySow,
		Handler: NewFarrowingHandler(
			createFarrowing, listFarrowings, updateFarrowing, deleteFarrowing, authMiddleware, adminMiddleware,
		),
	}
}
