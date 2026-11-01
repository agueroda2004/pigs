package infrastructure

import (
	"net/http"
	"time"

	serviceapplication "server/internal/modules/service/application"
	"server/internal/modules/service/ports"
)

type Module struct {
	CreateService          *serviceapplication.CreateServiceService
	ListServices           *serviceapplication.ListServicesService
	DeleteService          *serviceapplication.DeleteServiceService
	UpdateService          *serviceapplication.UpdateServiceService
	FarrowingSchedule       *serviceapplication.GetFarrowingScheduleService
	ServicesHistoryBySow   *serviceapplication.ServicesHistoryBySowService
	Handler                *ServiceHandler
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
	deleteService := serviceapplication.NewDeleteServiceService(repository, clock)
	updateService := serviceapplication.NewUpdateServiceService(repository, clock)
	farrowingSchedule := serviceapplication.NewGetFarrowingScheduleService(repository, clock)
	servicesHistoryBySow := serviceapplication.NewServicesHistoryBySowService(repository)

	return &Module{
		CreateService:          createService,
		ListServices:           listServices,
		DeleteService:          deleteService,
		UpdateService:          updateService,
		FarrowingSchedule:       farrowingSchedule,
		ServicesHistoryBySow:    servicesHistoryBySow,
		Handler:                NewServiceHandler(createService, listServices, deleteService, updateService, farrowingSchedule, authMiddleware, adminMiddleware),
	}
}
