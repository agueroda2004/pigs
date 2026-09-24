package infrastructure

import (
	"net/http"
	"time"

	sowapplication "server/internal/modules/sow/application"
	"server/internal/modules/sow/ports"
)

type Module struct {
	CreateSow      *sowapplication.CreateSowService
	ListSows       *sowapplication.ListSowsService
	UpdateSow      *sowapplication.UpdateSowService
	ChangeSowState *sowapplication.ChangeSowStateService
	Handler        *SowHandler
}

// NewModule assembles the sow use cases and HTTP handler from its dependencies.
// It returns a module exposing the services and handler for wiring; state changes
// stay internal and are not registered as HTTP routes.
func NewModule(
	repository ports.SowRepository,
	clock func() time.Time,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *Module {
	createSow := sowapplication.NewCreateSowService(repository, clock)
	listSows := sowapplication.NewListSowsService(repository)
	updateSow := sowapplication.NewUpdateSowService(repository, clock)
	changeSowState := sowapplication.NewChangeSowStateService(repository, clock)

	return &Module{
		CreateSow:      createSow,
		ListSows:       listSows,
		UpdateSow:      updateSow,
		ChangeSowState: changeSowState,
		Handler:        NewSowHandler(createSow, listSows, updateSow, authMiddleware, adminMiddleware),
	}
}
