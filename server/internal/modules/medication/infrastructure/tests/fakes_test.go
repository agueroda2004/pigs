package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	medicationapplication "server/internal/modules/medication/application"
	medicationdomain "server/internal/modules/medication/domain"
	medicationinfra "server/internal/modules/medication/infrastructure"
	"server/internal/modules/medication/ports"
	userdomain "server/internal/modules/user/domain"
)

type fakeCreateMedicationUseCase struct {
	medication *medicationdomain.Medication
	err        error
	command    medicationapplication.CreateMedicationCommand
	called     bool
}

func (f *fakeCreateMedicationUseCase) Execute(_ context.Context, command medicationapplication.CreateMedicationCommand) (*medicationdomain.Medication, error) {
	f.called = true
	f.command = command
	return f.medication, f.err
}

type fakeListMedicationsUseCase struct {
	medications []*medicationdomain.Medication
	err         error
	filter      ports.MedicationFilter
	called      bool
}

func (f *fakeListMedicationsUseCase) Execute(_ context.Context, filter ports.MedicationFilter) ([]*medicationdomain.Medication, error) {
	f.called = true
	f.filter = filter
	return f.medications, f.err
}

type fakeListMedicationOptionsUseCase struct {
	options []medicationdomain.MedicationOption
	err     error
	active  *bool
	called  bool
}

func (f *fakeListMedicationOptionsUseCase) Execute(_ context.Context, active *bool) ([]medicationdomain.MedicationOption, error) {
	f.called = true
	f.active = active
	return f.options, f.err
}

type fakeUpdateMedicationUseCase struct {
	medication   *medicationdomain.Medication
	err          error
	medicationID uuid.UUID
	command      medicationapplication.UpdateMedicationCommand
	called       bool
}

func (f *fakeUpdateMedicationUseCase) Execute(_ context.Context, medicationID uuid.UUID, command medicationapplication.UpdateMedicationCommand) (*medicationdomain.Medication, error) {
	f.called = true
	f.medicationID = medicationID
	f.command = command
	return f.medication, f.err
}

func newTestHandler(create medicationinfra.CreateMedicationUseCase, list medicationinfra.ListMedicationsUseCase, update medicationinfra.UpdateMedicationUseCase) *medicationinfra.MedicationHandler {
	return newTestHandlerWithOptions(create, list, &fakeListMedicationOptionsUseCase{}, update)
}

func newTestHandlerWithOptions(
	create medicationinfra.CreateMedicationUseCase,
	list medicationinfra.ListMedicationsUseCase,
	options medicationinfra.ListMedicationOptionsUseCase,
	update medicationinfra.UpdateMedicationUseCase,
) *medicationinfra.MedicationHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return medicationinfra.NewMedicationHandler(create, list, options, update, passThrough, passThrough)
}

func authenticatedRequest(request *http.Request, userID uuid.UUID) *http.Request {
	return request.WithContext(authdomain.WithAuthenticatedUser(request.Context(), authdomain.AuthenticatedUser{
		UserID: userID,
		Role:   userdomain.RoleAdmin,
	}))
}

func serve(handler *medicationinfra.MedicationHandler, request *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func handlerMedication() *medicationdomain.Medication {
	actor := uuid.New()
	return &medicationdomain.Medication{
		ID:        uuid.New(),
		Name:      "Ivermectina",
		Active:    true,
		CreatedAt: time.Date(2026, time.January, 10, 3, 4, 5, 0, time.UTC),
		UpdatedAt: time.Date(2026, time.January, 10, 3, 4, 5, 0, time.UTC),
		CreatedBy: actor,
		UpdatedBy: actor,
	}
}
