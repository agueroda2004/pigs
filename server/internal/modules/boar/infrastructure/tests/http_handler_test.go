package tests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	"server/internal/modules/boar/ports"
)

func TestBoarHandlerCreate(t *testing.T) {
	boar := handlerBoar()
	actorID := uuid.New()
	breedID := uuid.New()

	validBody := `{"code":"B-001","location":"Corral A","entry_date":"2026-01-10","birth_date":"2025-12-01","note":"ok","origin":"Propio","breed_id":"` + breedID.String() + `"}`

	t.Run("creates a boar with actor from context", func(t *testing.T) {
		create := &fakeCreateBoarUseCase{boar: boar}
		handler := newTestHandler(create, &fakeListBoarsUseCase{}, &fakeUpdateBoarUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/boars", strings.NewReader(validBody))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusCreated || create.command.CreatedBy != actorID || create.command.Code != "B-001" {
			t.Fatalf("status=%d command=%#v", response.Code, create.command)
		}
		if create.command.BreedID != breedID || create.command.EntryDate.IsZero() || create.command.BirthDate == nil {
			t.Fatalf("unexpected command: %#v", create.command)
		}
		if create.command.Origin != boardomain.OriginOwn {
			t.Fatalf("unexpected origin: %#v", create.command.Origin)
		}
		if response.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type: %q", response.Header().Get("Content-Type"))
		}
	})

	t.Run("rejects a state field", func(t *testing.T) {
		create := &fakeCreateBoarUseCase{boar: boar}
		handler := newTestHandler(create, &fakeListBoarsUseCase{}, &fakeUpdateBoarUseCase{})
		body := `{"code":"B-001","entry_date":"2026-01-10","breed_id":"` + breedID.String() + `","state":"Muerto"}`
		request := httptest.NewRequest(http.MethodPost, "/api/v1/boars", strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		create := &fakeCreateBoarUseCase{boar: boar}
		handler := newTestHandler(create, &fakeListBoarsUseCase{}, &fakeUpdateBoarUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/boars", strings.NewReader(validBody)))

		if response.Code != http.StatusUnauthorized || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid JSON", func(t *testing.T) {
		create := &fakeCreateBoarUseCase{boar: boar}
		handler := newTestHandler(create, &fakeListBoarsUseCase{}, &fakeUpdateBoarUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/boars", strings.NewReader(`{"code":`)))

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid dates and breed", func(t *testing.T) {
		create := &fakeCreateBoarUseCase{boar: boar}
		handler := newTestHandler(create, &fakeListBoarsUseCase{}, &fakeUpdateBoarUseCase{})

		for _, body := range []string{
			`{"code":"B-001","entry_date":"not-a-date","breed_id":"` + breedID.String() + `"}`,
			`{"code":"B-001","entry_date":"2026-01-10","birth_date":"not-a-date","breed_id":"` + breedID.String() + `"}`,
			`{"code":"B-001","entry_date":"2026-01-10","breed_id":"not-a-uuid"}`,
		} {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/boars", strings.NewReader(body))
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
			{"duplicate", ports.ErrBoarCodeAlreadyUsed, http.StatusConflict},
			{"validation", boardomain.ErrInvalidCode, http.StatusBadRequest},
			{"invalid origin", boardomain.ErrInvalidOrigin, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				create := &fakeCreateBoarUseCase{err: test.err}
				handler := newTestHandler(create, &fakeListBoarsUseCase{}, &fakeUpdateBoarUseCase{})
				request := httptest.NewRequest(http.MethodPost, "/api/v1/boars", strings.NewReader(validBody))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestBoarHandlerList(t *testing.T) {
	t.Run("returns every boar", func(t *testing.T) {
		list := &fakeListBoarsUseCase{boars: []*boardomain.Boar{handlerBoar(), handlerBoar()}}
		handler := newTestHandler(&fakeCreateBoarUseCase{}, list, &fakeUpdateBoarUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/boars", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		body := strings.TrimSpace(response.Body.String())
		if !strings.HasPrefix(body, "[") || !strings.Contains(body, "B-001") || !strings.Contains(body, "Vivo") || !strings.Contains(body, "Propio") {
			t.Fatalf("unexpected body: %s", body)
		}
	})

	t.Run("returns an empty array when there are no boars", func(t *testing.T) {
		list := &fakeListBoarsUseCase{boars: []*boardomain.Boar{}}
		handler := newTestHandler(&fakeCreateBoarUseCase{}, list, &fakeUpdateBoarUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/boars", nil))

		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		list := &fakeListBoarsUseCase{err: errors.New("unexpected")}
		handler := newTestHandler(&fakeCreateBoarUseCase{}, list, &fakeUpdateBoarUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/boars", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})

	t.Run("parses the query filters", func(t *testing.T) {
		breedID := uuid.New()
		list := &fakeListBoarsUseCase{}
		handler := newTestHandler(&fakeCreateBoarUseCase{}, list, &fakeUpdateBoarUseCase{})
		target := "/api/v1/boars?code=%20B-001%20&breed_id=" + breedID.String() + "&origin=Externo&active=false"
		response := serve(handler, httptest.NewRequest(http.MethodGet, target, nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		filter := list.filter
		if filter.Code == nil || *filter.Code != "B-001" {
			t.Fatalf("unexpected code: %#v", filter.Code)
		}
		if filter.BreedID == nil || *filter.BreedID != breedID {
			t.Fatalf("unexpected breed id: %#v", filter.BreedID)
		}
		if filter.Origin == nil || *filter.Origin != boardomain.OriginExternal {
			t.Fatalf("unexpected origin: %#v", filter.Origin)
		}
		if filter.Active == nil || *filter.Active {
			t.Fatalf("unexpected active: %#v", filter.Active)
		}
	})

	t.Run("returns an empty filter when no params are given", func(t *testing.T) {
		list := &fakeListBoarsUseCase{}
		handler := newTestHandler(&fakeCreateBoarUseCase{}, list, &fakeUpdateBoarUseCase{})
		serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/boars", nil))

		filter := list.filter
		if filter.Code != nil || filter.BreedID != nil || filter.Origin != nil || filter.Active != nil {
			t.Fatalf("unexpected filter: %#v", filter)
		}
	})

	t.Run("rejects invalid filter values", func(t *testing.T) {
		list := &fakeListBoarsUseCase{}
		handler := newTestHandler(&fakeCreateBoarUseCase{}, list, &fakeUpdateBoarUseCase{})

		for _, target := range []string{
			"/api/v1/boars?breed_id=not-a-uuid",
			"/api/v1/boars?origin=Desconocido",
			"/api/v1/boars?active=maybe",
		} {
			response := serve(handler, httptest.NewRequest(http.MethodGet, target, nil))
			if response.Code != http.StatusBadRequest || list.called {
				t.Fatalf("target=%s status=%d called=%v", target, response.Code, list.called)
			}
		}
	})
}

func TestBoarHandlerUpdate(t *testing.T) {
	boarID := uuid.New()
	actorID := uuid.New()
	breedID := uuid.New()

	t.Run("updates the provided fields and forwards the actor", func(t *testing.T) {
		update := &fakeUpdateBoarUseCase{boar: handlerBoar()}
		handler := newTestHandler(&fakeCreateBoarUseCase{}, &fakeListBoarsUseCase{}, update)
		body := `{"code":"B-002","active":false,"entry_date":"2026-02-01","breed_id":"` + breedID.String() + `"}`
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/boars/"+boarID.String(), strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || update.boarID != boarID || update.command.UpdatedBy != actorID {
			t.Fatalf("status=%d id=%v command=%#v", response.Code, update.boarID, update.command)
		}
		if update.command.Code == nil || *update.command.Code != "B-002" || update.command.Active == nil || *update.command.Active {
			t.Fatalf("unexpected command: %#v", update.command)
		}
		if update.command.EntryDate == nil || update.command.BreedID == nil || *update.command.BreedID != breedID {
			t.Fatalf("unexpected command: %#v", update.command)
		}
	})

	t.Run("forwards the origin", func(t *testing.T) {
		update := &fakeUpdateBoarUseCase{boar: handlerBoar()}
		handler := newTestHandler(&fakeCreateBoarUseCase{}, &fakeListBoarsUseCase{}, update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/boars/"+boarID.String(), strings.NewReader(`{"origin":"Externo"}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || !update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
		if update.command.Origin == nil || *update.command.Origin != boardomain.OriginExternal {
			t.Fatalf("unexpected origin: %#v", update.command.Origin)
		}
	})

	t.Run("forwards empty nullable fields to clear them", func(t *testing.T) {
		update := &fakeUpdateBoarUseCase{boar: handlerBoar()}
		handler := newTestHandler(&fakeCreateBoarUseCase{}, &fakeListBoarsUseCase{}, update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/boars/"+boarID.String(), strings.NewReader(`{"location":"","note":"","birth_date":""}`))
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
		update := &fakeUpdateBoarUseCase{boar: handlerBoar()}
		handler := newTestHandler(&fakeCreateBoarUseCase{}, &fakeListBoarsUseCase{}, update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/boars/"+boarID.String(), strings.NewReader(`{"state":"Muerto"}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("returns bad request for invalid UUID", func(t *testing.T) {
		update := &fakeUpdateBoarUseCase{}
		handler := newTestHandler(&fakeCreateBoarUseCase{}, &fakeListBoarsUseCase{}, update)
		response := serve(handler, httptest.NewRequest(http.MethodPatch, "/api/v1/boars/not-a-uuid", strings.NewReader(`{"code":"B-002"}`)))

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("returns bad request for invalid date and breed", func(t *testing.T) {
		update := &fakeUpdateBoarUseCase{boar: handlerBoar()}
		handler := newTestHandler(&fakeCreateBoarUseCase{}, &fakeListBoarsUseCase{}, update)

		for _, body := range []string{
			`{"entry_date":"not-a-date"}`,
			`{"birth_date":"not-a-date"}`,
			`{"breed_id":"not-a-uuid"}`,
		} {
			request := httptest.NewRequest(http.MethodPatch, "/api/v1/boars/"+boarID.String(), strings.NewReader(body))
			request = authenticatedRequest(request, actorID)
			response := serve(handler, request)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", response.Code, body)
			}
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		update := &fakeUpdateBoarUseCase{boar: handlerBoar()}
		handler := newTestHandler(&fakeCreateBoarUseCase{}, &fakeListBoarsUseCase{}, update)
		response := serve(handler, httptest.NewRequest(http.MethodPatch, "/api/v1/boars/"+boarID.String(), strings.NewReader(`{"code":"B-002"}`)))

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
			{"not found", ports.ErrBoarNotFound, http.StatusNotFound},
			{"duplicate", ports.ErrBoarCodeAlreadyUsed, http.StatusConflict},
			{"validation", boardomain.ErrInvalidUpdate, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				update := &fakeUpdateBoarUseCase{err: test.err}
				handler := newTestHandler(&fakeCreateBoarUseCase{}, &fakeListBoarsUseCase{}, update)
				request := httptest.NewRequest(http.MethodPatch, "/api/v1/boars/"+uuid.New().String(), strings.NewReader(`{"code":"B-002"}`))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}
