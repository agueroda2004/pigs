package infrastructure

import (
	"net/http"
	"time"

	sowremovalapplication "server/internal/modules/sowremoval/application"
	"server/internal/modules/sowremoval/ports"
)

type Module struct {
	CreateSowRemoval *sowremovalapplication.CreateSowRemovalService
	ListSowRemovals  *sowremovalapplication.ListSowRemovalsService
	Handler          *SowRemovalHandler
}

// NewModule assembles the removal use cases and HTTP handler from its dependencies.
// It returns a module exposing the use cases and handler for wiring.
func NewModule(
	repository ports.SowRemovalRepository,
	clock func() time.Time,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *Module {
	createSowRemoval := sowremovalapplication.NewCreateSowRemovalService(repository, clock)
	listSowRemovals := sowremovalapplication.NewListSowRemovalsService(repository)

	return &Module{
		CreateSowRemoval: createSowRemoval,
		ListSowRemovals:  listSowRemovals,
		Handler:          NewSowRemovalHandler(createSowRemoval, listSowRemovals, authMiddleware, adminMiddleware),
	}
}
