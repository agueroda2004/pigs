package tests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	sowdomain "server/internal/modules/sow/domain"
	"server/internal/modules/sow/ports"
)

func TestSowHandlerCreate(t *testing.T) {
	sow := handlerSow()
	actorID := uuid.New()
	breedID := uuid.New()

	validBody := `{"code":"C-001","location":"Corral A","entry_date":"2026-01-10","birth_date":"2025-12-01","note":"ok","origin":"Propio","parity":3,"breed_id":"` + breedID.String() + `"}`

	t.Run("creates a sow with actor from context", func(t *testing.T) {
		create := &fakeCreateSowUseCase{sow: sow}
		handler := newTestHandler(create, &fakeListSowsUseCase{}, &fakeUpdateSowUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/sows", strings.NewReader(validBody))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusCreated || create.command.CreatedBy != actorID || create.command.Code != "C-001" {
			t.Fatalf("status=%d command=%#v", response.Code, create.command)
		}
		if create.command.BreedID != breedID || create.command.EntryDate.IsZero() || create.command.BirthDate == nil {
			t.Fatalf("unexpected command: %#v", create.command)
		}
		if create.command.Origin != sowdomain.OriginOwn || create.command.Parity != 3 {
			t.Fatalf("unexpected command: %#v", create.command)
		}
		if response.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type: %q", response.Header().Get("Content-Type"))
		}
	})

	t.Run("rejects a state field", func(t *testing.T) {
		create := &fakeCreateSowUseCase{sow: sow}
		handler := newTestHandler(create, &fakeListSowsUseCase{}, &fakeUpdateSowUseCase{})
		body := `{"code":"C-001","entry_date":"2026-01-10","breed_id":"` + breedID.String() + `","state":"Muerta"}`
		request := httptest.NewRequest(http.MethodPost, "/api/v1/sows", strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		create := &fakeCreateSowUseCase{sow: sow}
		handler := newTestHandler(create, &fakeListSowsUseCase{}, &fakeUpdateSowUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/sows", strings.NewReader(validBody)))

		if response.Code != http.StatusUnauthorized || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid JSON", func(t *testing.T) {
		create := &fakeCreateSowUseCase{sow: sow}
		handler := newTestHandler(create, &fakeListSowsUseCase{}, &fakeUpdateSowUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/sows", strings.NewReader(`{"code":`)))

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid dates and breed", func(t *testing.T) {
		create := &fakeCreateSowUseCase{sow: sow}
		handler := newTestHandler(create, &fakeListSowsUseCase{}, &fakeUpdateSowUseCase{})

		for _, body := range []string{
			`{"code":"C-001","entry_date":"not-a-date","breed_id":"` + breedID.String() + `"}`,
			`{"code":"C-001","entry_date":"2026-01-10","birth_date":"not-a-date","breed_id":"` + breedID.String() + `"}`,
			`{"code":"C-001","entry_date":"2026-01-10","breed_id":"not-a-uuid"}`,
		} {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/sows", strings.NewReader(body))
			request = authenticatedRequest(request, actorID)
			response := serve(handler, request)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", response.Code, body)
			}
		}
	})

	t.Run("maps application errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"duplicate", ports.ErrSowCodeAlreadyUsed, http.StatusConflict},
			{"validation", sowdomain.ErrInvalidCode, http.StatusBadRequest},
			{"invalid parity", sowdomain.ErrInvalidParity, http.StatusBadRequest},
			{"invalid origin", sowdomain.ErrInvalidOrigin, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				create := &fakeCreateSowUseCase{err: test.err}
				handler := newTestHandler(create, &fakeListSowsUseCase{}, &fakeUpdateSowUseCase{})
				request := httptest.NewRequest(http.MethodPost, "/api/v1/sows", strings.NewReader(validBody))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestSowHandlerList(t *testing.T) {
	t.Run("returns every sow", func(t *testing.T) {
		list := &fakeListSowsUseCase{sows: []*sowdomain.Sow{handlerSow(), handlerSow()}}
		handler := newTestHandler(&fakeCreateSowUseCase{}, list, &fakeUpdateSowUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		body := strings.TrimSpace(response.Body.String())
		if !strings.HasPrefix(body, "[") || !strings.Contains(body, "C-001") || !strings.Contains(body, "Viva") || !strings.Contains(body, "Propio") {
			t.Fatalf("unexpected body: %s", body)
		}
	})

	t.Run("returns an empty array when there are no sows", func(t *testing.T) {
		list := &fakeListSowsUseCase{sows: []*sowdomain.Sow{}}
		handler := newTestHandler(&fakeCreateSowUseCase{}, list, &fakeUpdateSowUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows", nil))

		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		list := &fakeListSowsUseCase{err: errors.New("unexpected")}
		handler := newTestHandler(&fakeCreateSowUseCase{}, list, &fakeUpdateSowUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})

	t.Run("parses the query filters", func(t *testing.T) {
		breedID := uuid.New()
		list := &fakeListSowsUseCase{}
		handler := newTestHandler(&fakeCreateSowUseCase{}, list, &fakeUpdateSowUseCase{})
		target := "/api/v1/sows?code=%20C-001%20&breed_id=" + breedID.String() + "&origin=Externo&active=false&state=Gestando"
		response := serve(handler, httptest.NewRequest(http.MethodGet, target, nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		filter := list.filter
		if filter.Code == nil || *filter.Code != "C-001" {
			t.Fatalf("unexpected code: %#v", filter.Code)
		}
		if filter.BreedID == nil || *filter.BreedID != breedID {
			t.Fatalf("unexpected breed id: %#v", filter.BreedID)
		}
		if filter.Origin == nil || *filter.Origin != sowdomain.OriginExternal {
			t.Fatalf("unexpected origin: %#v", filter.Origin)
		}
		if filter.Active == nil || *filter.Active {
			t.Fatalf("unexpected active: %#v", filter.Active)
		}
		if filter.State == nil || *filter.State != sowdomain.StatePregnant {
			t.Fatalf("unexpected state: %#v", filter.State)
		}
	})

	t.Run("returns an empty filter when no params are given", func(t *testing.T) {
		list := &fakeListSowsUseCase{}
		handler := newTestHandler(&fakeCreateSowUseCase{}, list, &fakeUpdateSowUseCase{})
		serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows", nil))

		filter := list.filter
		if filter.Code != nil || filter.BreedID != nil || filter.Origin != nil || filter.Active != nil || filter.State != nil {
			t.Fatalf("unexpected filter: %#v", filter)
		}
	})

	t.Run("rejects invalid filter values", func(t *testing.T) {
		list := &fakeListSowsUseCase{}
		handler := newTestHandler(&fakeCreateSowUseCase{}, list, &fakeUpdateSowUseCase{})

		for _, target := range []string{
			"/api/v1/sows?breed_id=not-a-uuid",
			"/api/v1/sows?origin=Desconocido",
			"/api/v1/sows?active=maybe",
			"/api/v1/sows?state=Desconocido",
		} {
			response := serve(handler, httptest.NewRequest(http.MethodGet, target, nil))
			if response.Code != http.StatusBadRequest || list.called {
				t.Fatalf("target=%s status=%d called=%v", target, response.Code, list.called)
			}
		}
	})
}

func TestSowHandlerUpdate(t *testing.T) {
	sowID := uuid.New()
	actorID := uuid.New()
	breedID := uuid.New()

	t.Run("updates the provided fields and forwards the actor", func(t *testing.T) {
		update := &fakeUpdateSowUseCase{sow: handlerSow()}
		handler := newTestHandler(&fakeCreateSowUseCase{}, &fakeListSowsUseCase{}, update)
		body := `{"code":"C-002","active":false,"entry_date":"2026-02-01","breed_id":"` + breedID.String() + `"}`
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/sows/"+sowID.String(), strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || update.sowID != sowID || update.command.UpdatedBy != actorID {
			t.Fatalf("status=%d id=%v command=%#v", response.Code, update.sowID, update.command)
		}
		if update.command.Code == nil || *update.command.Code != "C-002" || update.command.Active == nil || *update.command.Active {
			t.Fatalf("unexpected command: %#v", update.command)
		}
		if update.command.EntryDate == nil || update.command.BreedID == nil || *update.command.BreedID != breedID {
			t.Fatalf("unexpected command: %#v", update.command)
		}
	})

	t.Run("forwards the origin", func(t *testing.T) {
		update := &fakeUpdateSowUseCase{sow: handlerSow()}
		handler := newTestHandler(&fakeCreateSowUseCase{}, &fakeListSowsUseCase{}, update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/sows/"+sowID.String(), strings.NewReader(`{"origin":"Externo"}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || !update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
		if update.command.Origin == nil || *update.command.Origin != sowdomain.OriginExternal {
			t.Fatalf("unexpected origin: %#v", update.command.Origin)
		}
	})

	t.Run("forwards empty nullable fields to clear them", func(t *testing.T) {
		update := &fakeUpdateSowUseCase{sow: handlerSow()}
		handler := newTestHandler(&fakeCreateSowUseCase{}, &fakeListSowsUseCase{}, update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/sows/"+sowID.String(), strings.NewReader(`{"location":"","note":"","birth_date":""}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || !update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
		if update.command.Location == nil || *update.command.Location != "" {
			t.Fatalf("unexpected location: %#v", update.command.Location)
		}
		if update.command.Note == nil || *update.command.Note != "" {
			t.Fatalf("unexpected note: %#v", update.command.Note)
		}
		if !update.command.ClearBirthDate || update.command.BirthDate != nil {
			t.Fatalf("unexpected birth date: %#v", update.command)
		}
	})

	t.Run("rejects a state field", func(t *testing.T) {
		update := &fakeUpdateSowUseCase{sow: handlerSow()}
		handler := newTestHandler(&fakeCreateSowUseCase{}, &fakeListSowsUseCase{}, update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/sows/"+sowID.String(), strings.NewReader(`{"state":"Muerta"}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("rejects a parity field", func(t *testing.T) {
		update := &fakeUpdateSowUseCase{sow: handlerSow()}
		handler := newTestHandler(&fakeCreateSowUseCase{}, &fakeListSowsUseCase{}, update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/sows/"+sowID.String(), strings.NewReader(`{"parity":5}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("returns bad request for invalid UUID", func(t *testing.T) {
		update := &fakeUpdateSowUseCase{}
		handler := newTestHandler(&fakeCreateSowUseCase{}, &fakeListSowsUseCase{}, update)
		response := serve(handler, httptest.NewRequest(http.MethodPatch, "/api/v1/sows/not-a-uuid", strings.NewReader(`{"code":"C-002"}`)))

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("returns bad request for invalid date and breed", func(t *testing.T) {
		update := &fakeUpdateSowUseCase{sow: handlerSow()}
		handler := newTestHandler(&fakeCreateSowUseCase{}, &fakeListSowsUseCase{}, update)

		for _, body := range []string{
			`{"entry_date":"not-a-date"}`,
			`{"birth_date":"not-a-date"}`,
			`{"breed_id":"not-a-uuid"}`,
		} {
			request := httptest.NewRequest(http.MethodPatch, "/api/v1/sows/"+sowID.String(), strings.NewReader(body))
			request = authenticatedRequest(request, actorID)
			response := serve(handler, request)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", response.Code, body)
			}
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		update := &fakeUpdateSowUseCase{sow: handlerSow()}
		handler := newTestHandler(&fakeCreateSowUseCase{}, &fakeListSowsUseCase{}, update)
		response := serve(handler, httptest.NewRequest(http.MethodPatch, "/api/v1/sows/"+sowID.String(), strings.NewReader(`{"code":"C-002"}`)))

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
			{"not found", ports.ErrSowNotFound, http.StatusNotFound},
			{"duplicate", ports.ErrSowCodeAlreadyUsed, http.StatusConflict},
			{"validation", sowdomain.ErrInvalidUpdate, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				update := &fakeUpdateSowUseCase{err: test.err}
				handler := newTestHandler(&fakeCreateSowUseCase{}, &fakeListSowsUseCase{}, update)
				request := httptest.NewRequest(http.MethodPatch, "/api/v1/sows/"+uuid.New().String(), strings.NewReader(`{"code":"C-002"}`))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}
