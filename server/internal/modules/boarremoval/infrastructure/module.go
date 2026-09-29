package infrastructure

import (
	"net/http"
	"time"

	boarremovalapplication "server/internal/modules/boarremoval/application"
	"server/internal/modules/boarremoval/ports"
)

type Module struct {
	CreateBoarRemoval *boarremovalapplication.CreateBoarRemovalService
	UpdateBoarRemoval *boarremovalapplication.UpdateBoarRemovalService
	DeleteBoarRemoval *boarremovalapplication.DeleteBoarRemovalService
	ListBoarRemovals  *boarremovalapplication.ListBoarRemovalsService
	Handler           *BoarRemovalHandler
}

// NewModule assembles the removal use cases and HTTP handler from its dependencies.
// It returns a module exposing the use cases and handler for wiring.
func NewModule(
	repository ports.BoarRemovalRepository,
	clock func() time.Time,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *Module {
	createBoarRemoval := boarremovalapplication.NewCreateBoarRemovalService(repository, clock)
	updateBoarRemoval := boarremovalapplication.NewUpdateBoarRemovalService(repository, clock)
	deleteBoarRemoval := boarremovalapplication.NewDeleteBoarRemovalService(repository, clock)
	listBoarRemovals := boarremovalapplication.NewListBoarRemovalsService(repository)

	return &Module{
		CreateBoarRemoval: createBoarRemoval,
		UpdateBoarRemoval: updateBoarRemoval,
		DeleteBoarRemoval: deleteBoarRemoval,
		ListBoarRemovals:  listBoarRemovals,
		Handler:           NewBoarRemovalHandler(createBoarRemoval, updateBoarRemoval, deleteBoarRemoval, listBoarRemovals, authMiddleware, adminMiddleware),
	}
}
