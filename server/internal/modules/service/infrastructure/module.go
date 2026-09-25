package infrastructure

import (
	"net/http"
	"time"

	serviceapplication "server/internal/modules/service/application"
	"server/internal/modules/service/ports"
)

type Module struct {
	CreateService *serviceapplication.CreateServiceService
	ListServices  *serviceapplication.ListServicesService
	Handler       *ServiceHandler
}

// NewModule assembles the service use cases and HTTP handler from its dependencies.
// It returns a module exposing the services and handler for wiring.
func NewModule(
	repository ports.ServiceRepository,
	clock func() time.Time,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *Module {
	createService := serviceapplication.NewCreateServiceService(repository, clock)
	listServices := serviceapplication.NewListServicesService(repository)

	return &Module{
		CreateService: createService,
		ListServices:  listServices,
		Handler:       NewServiceHandler(createService, listServices, authMiddleware, adminMiddleware),
	}
}
