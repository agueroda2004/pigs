package infrastructure

import (
	"net/http"
	"time"

	sowapplication "server/internal/modules/sow/application"
	"server/internal/modules/sow/ports"
)

type Module struct {
	CreateSow       *sowapplication.CreateSowService
	ListSows        *sowapplication.ListSowsService
	ListSowDropdown *sowapplication.ListSowDropdownService
	UpdateSow       *sowapplication.UpdateSowService
	DeleteSow       *sowapplication.DeleteSowService
	ChangeSowState  *sowapplication.ChangeSowStateService
	Handler         *SowHandler
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
	listSowDropdown := sowapplication.NewListSowDropdownService(repository)
	updateSow := sowapplication.NewUpdateSowService(repository, clock)
	deleteSow := sowapplication.NewDeleteSowService(repository)
	changeSowState := sowapplication.NewChangeSowStateService(repository, clock)

	return &Module{
		CreateSow:       createSow,
		ListSows:        listSows,
		ListSowDropdown: listSowDropdown,
		UpdateSow:       updateSow,
		DeleteSow:       deleteSow,
		ChangeSowState:  changeSowState,
		Handler:         NewSowHandler(createSow, listSows, listSowDropdown, updateSow, deleteSow, authMiddleware, adminMiddleware),
	}
}
