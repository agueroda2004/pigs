package infrastructure

import (
	"net/http"
	"time"

	medicationapplication "server/internal/modules/medication/application"
	"server/internal/modules/medication/ports"
)

type Module struct {
	CreateMedication      *medicationapplication.CreateMedicationService
	ListMedications       *medicationapplication.ListMedicationsService
	ListMedicationOptions *medicationapplication.ListMedicationOptionsService
	UpdateMedication      *medicationapplication.UpdateMedicationService
	Handler               *MedicationHandler
}

// NewModule assembles the medication use cases and HTTP handler from its dependencies.
// It returns a module exposing the services and handler for wiring.
func NewModule(
	repository ports.MedicationRepository,
	clock func() time.Time,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *Module {
	createMedication := medicationapplication.NewCreateMedicationService(repository, clock)
	listMedications := medicationapplication.NewListMedicationsService(repository)
	listMedicationOptions := medicationapplication.NewListMedicationOptionsService(repository)
	updateMedication := medicationapplication.NewUpdateMedicationService(repository, clock)

	return &Module{
		CreateMedication:      createMedication,
		ListMedications:       listMedications,
		ListMedicationOptions: listMedicationOptions,
		UpdateMedication:      updateMedication,
		Handler:               NewMedicationHandler(createMedication, listMedications, listMedicationOptions, updateMedication, authMiddleware, adminMiddleware),
	}
}
