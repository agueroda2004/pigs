package tests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	medicationdomain "server/internal/modules/medication/domain"
	medicationinfra "server/internal/modules/medication/infrastructure"
	"server/internal/modules/medication/ports"
)

func TestMedicationHandlerCreate(t *testing.T) {
	medication := handlerMedication()
	actorID := uuid.New()

	t.Run("creates a medication with actor from context", func(t *testing.T) {
		create := &fakeCreateMedicationUseCase{medication: medication}
		handler := newTestHandler(create, &fakeListMedicationsUseCase{}, &fakeUpdateMedicationUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/medications", strings.NewReader(`{"name":"Ivermectina"}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusCreated || create.command.CreatedBy != actorID || create.command.Name != "Ivermectina" {
			t.Fatalf("status=%d command=%#v", response.Code, create.command)
		}
		if response.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type: %q", response.Header().Get("Content-Type"))
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		create := &fakeCreateMedicationUseCase{medication: medication}
		handler := newTestHandler(create, &fakeListMedicationsUseCase{}, &fakeUpdateMedicationUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/medications", strings.NewReader(`{"name":"Ivermectina"}`)))

		if response.Code != http.StatusUnauthorized || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid JSON", func(t *testing.T) {
		create := &fakeCreateMedicationUseCase{medication: medication}
		handler := newTestHandler(create, &fakeListMedicationsUseCase{}, &fakeUpdateMedicationUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/medications", strings.NewReader(`{"name":`)))

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("maps application errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"duplicate", ports.ErrMedicationNameAlreadyUsed, http.StatusConflict},
			{"validation", medicationdomain.ErrInvalidName, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				create := &fakeCreateMedicationUseCase{err: test.err}
				handler := newTestHandler(create, &fakeListMedicationsUseCase{}, &fakeUpdateMedicationUseCase{})
				request := httptest.NewRequest(http.MethodPost, "/api/v1/medications", strings.NewReader(`{"name":"Ivermectina"}`))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestMedicationHandlerList(t *testing.T) {
	t.Run("forwards the name and active filters", func(t *testing.T) {
		list := &fakeListMedicationsUseCase{medications: []*medicationdomain.Medication{handlerMedication()}}
		handler := newTestHandler(&fakeCreateMedicationUseCase{}, list, &fakeUpdateMedicationUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/medications?name=%20Iver%20&active=false", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		if list.filter.Name == nil || *list.filter.Name != "Iver" {
			t.Fatalf("unexpected name filter: %#v", list.filter)
		}
		if list.filter.Active == nil || *list.filter.Active {
			t.Fatalf("unexpected active filter: %#v", list.filter)
		}
	})

	t.Run("returns bad request for an invalid active value", func(t *testing.T) {
		list := &fakeListMedicationsUseCase{}
		handler := newTestHandler(&fakeCreateMedicationUseCase{}, list, &fakeUpdateMedicationUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/medications?active=maybe", nil))

		if response.Code != http.StatusBadRequest || list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
	})

	t.Run("returns an empty array when there are no medications", func(t *testing.T) {
		list := &fakeListMedicationsUseCase{medications: []*medicationdomain.Medication{}}
		handler := newTestHandler(&fakeCreateMedicationUseCase{}, list, &fakeUpdateMedicationUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/medications", nil))

		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		list := &fakeListMedicationsUseCase{err: errors.New("unexpected")}
		handler := newTestHandler(&fakeCreateMedicationUseCase{}, list, &fakeUpdateMedicationUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/medications", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}

func TestMedicationHandlerListOptions(t *testing.T) {
	newOptionsHandler := func(options *fakeListMedicationOptionsUseCase) *medicationinfra.MedicationHandler {
		return newTestHandlerWithOptions(
			&fakeCreateMedicationUseCase{},
			&fakeListMedicationsUseCase{},
			options,
			&fakeUpdateMedicationUseCase{},
		)
	}

	t.Run("returns only active options by default", func(t *testing.T) {
		options := &fakeListMedicationOptionsUseCase{
			options: []medicationdomain.MedicationOption{
				{ID: uuid.New(), Name: "Ivermectina"},
				{ID: uuid.New(), Name: "Penicilina"},
			},
		}
		handler := newOptionsHandler(options)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/medications/options", nil))

		if response.Code != http.StatusOK || !options.called {
			t.Fatalf("status=%d called=%v", response.Code, options.called)
		}
		if options.active == nil || !*options.active {
			t.Fatalf("unexpected active filter: %#v", options.active)
		}
		body := response.Body.String()
		if !strings.Contains(body, "Ivermectina") || !strings.Contains(body, "Penicilina") {
			t.Fatalf("unexpected body: %s", body)
		}
	})

	t.Run("includes inactive when include_inactive is true", func(t *testing.T) {
		options := &fakeListMedicationOptionsUseCase{options: []medicationdomain.MedicationOption{}}
		handler := newOptionsHandler(options)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/medications/options?include_inactive=true", nil))

		if response.Code != http.StatusOK || options.active != nil {
			t.Fatalf("status=%d active=%#v", response.Code, options.active)
		}
	})

	t.Run("keeps active-only when include_inactive is false", func(t *testing.T) {
		options := &fakeListMedicationOptionsUseCase{options: []medicationdomain.MedicationOption{}}
		handler := newOptionsHandler(options)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/medications/options?include_inactive=false", nil))

		if response.Code != http.StatusOK || options.active == nil || !*options.active {
			t.Fatalf("status=%d active=%#v", response.Code, options.active)
		}
	})

	t.Run("rejects an invalid include_inactive value", func(t *testing.T) {
		options := &fakeListMedicationOptionsUseCase{}
		handler := newOptionsHandler(options)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/medications/options?include_inactive=maybe", nil))

		if response.Code != http.StatusBadRequest || options.called {
			t.Fatalf("status=%d called=%v", response.Code, options.called)
		}
	})

	t.Run("returns an empty array when there are no options", func(t *testing.T) {
		options := &fakeListMedicationOptionsUseCase{options: []medicationdomain.MedicationOption{}}
		handler := newOptionsHandler(options)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/medications/options", nil))

		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		options := &fakeListMedicationOptionsUseCase{err: errors.New("unexpected")}
		handler := newOptionsHandler(options)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/medications/options", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}

func TestMedicationHandlerUpdate(t *testing.T) {
	medicationID := uuid.New()
	actorID := uuid.New()

	t.Run("updates the provided fields and forwards the actor", func(t *testing.T) {
		update := &fakeUpdateMedicationUseCase{medication: handlerMedication()}
		handler := newTestHandler(&fakeCreateMedicationUseCase{}, &fakeListMedicationsUseCase{}, update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/medications/"+medicationID.String(), strings.NewReader(`{"name":"New Name","active":false}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || update.medicationID != medicationID || update.command.UpdatedBy != actorID || update.command.Name == nil || *update.command.Name != "New Name" || update.command.Active == nil || *update.command.Active {
			t.Fatalf("status=%d id=%v command=%#v", response.Code, update.medicationID, update.command)
		}
	})

	t.Run("returns bad request for invalid UUID", func(t *testing.T) {
		update := &fakeUpdateMedicationUseCase{}
		handler := newTestHandler(&fakeCreateMedicationUseCase{}, &fakeListMedicationsUseCase{}, update)
		response := serve(handler, httptest.NewRequest(http.MethodPatch, "/api/v1/medications/not-a-uuid", strings.NewReader(`{"name":"New"}`)))

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		update := &fakeUpdateMedicationUseCase{medication: handlerMedication()}
		handler := newTestHandler(&fakeCreateMedicationUseCase{}, &fakeListMedicationsUseCase{}, update)
		response := serve(handler, httptest.NewRequest(http.MethodPatch, "/api/v1/medications/"+medicationID.String(), strings.NewReader(`{"name":"New"}`)))

		if response.Code != http.StatusUnauthorized || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("maps application errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"not found", ports.ErrMedicationNotFound, http.StatusNotFound},
			{"duplicate", ports.ErrMedicationNameAlreadyUsed, http.StatusConflict},
			{"validation", medicationdomain.ErrInvalidUpdate, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				update := &fakeUpdateMedicationUseCase{err: test.err}
				handler := newTestHandler(&fakeCreateMedicationUseCase{}, &fakeListMedicationsUseCase{}, update)
				request := httptest.NewRequest(http.MethodPatch, "/api/v1/medications/"+uuid.New().String(), strings.NewReader(`{"name":"New"}`))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}
