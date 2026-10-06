package tests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	breeddomain "server/internal/modules/breed/domain"
	"server/internal/modules/breed/ports"
)

func TestBreedHandlerCreate(t *testing.T) {
	breed := handlerBreed()
	actorID := uuid.New()

	t.Run("creates a breed with actor from context", func(t *testing.T) {
		create := &fakeCreateBreedUseCase{breed: breed}
		handler := newTestHandler(create, &fakeListBreedsUseCase{}, &fakeListBreedDropdownUseCase{}, &fakeUpdateBreedUseCase{}, &fakeDeleteBreedUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/breeds", strings.NewReader(`{"name":"Duroc"}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusCreated || create.command.CreatedBy != actorID || create.command.Name != "Duroc" {
			t.Fatalf("status=%d command=%#v", response.Code, create.command)
		}
		if body := strings.TrimSpace(response.Body.String()); body != "" {
			t.Fatalf("expected an empty body, got %q", body)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		create := &fakeCreateBreedUseCase{breed: breed}
		handler := newTestHandler(create, &fakeListBreedsUseCase{}, &fakeListBreedDropdownUseCase{}, &fakeUpdateBreedUseCase{}, &fakeDeleteBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/breeds", strings.NewReader(`{"name":"Duroc"}`)))

		if response.Code != http.StatusUnauthorized || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid JSON", func(t *testing.T) {
		create := &fakeCreateBreedUseCase{breed: breed}
		handler := newTestHandler(create, &fakeListBreedsUseCase{}, &fakeListBreedDropdownUseCase{}, &fakeUpdateBreedUseCase{}, &fakeDeleteBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/breeds", strings.NewReader(`{"name":`)))

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
			{"duplicate", ports.ErrBreedNameAlreadyUsed, http.StatusConflict},
			{"validation", breeddomain.ErrInvalidName, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				create := &fakeCreateBreedUseCase{err: test.err}
				handler := newTestHandler(create, &fakeListBreedsUseCase{}, &fakeListBreedDropdownUseCase{}, &fakeUpdateBreedUseCase{}, &fakeDeleteBreedUseCase{})
				request := httptest.NewRequest(http.MethodPost, "/api/v1/breeds", strings.NewReader(`{"name":"Duroc"}`))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestBreedHandlerList(t *testing.T) {
	t.Run("returns every breed without the audit fields", func(t *testing.T) {
		list := &fakeListBreedsUseCase{breeds: []*breeddomain.Breed{handlerBreed(), handlerBreed()}}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, list, &fakeListBreedDropdownUseCase{}, &fakeUpdateBreedUseCase{}, &fakeDeleteBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/breeds", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		body := strings.TrimSpace(response.Body.String())
		if !strings.HasPrefix(body, "[") || !strings.Contains(body, "Duroc") || !strings.Contains(body, "active") {
			t.Fatalf("unexpected body: %s", body)
		}
		if strings.Contains(body, "created_at") || strings.Contains(body, "created_by") {
			t.Fatalf("body must not include audit fields: %s", body)
		}
	})

	t.Run("parses the name and active filters", func(t *testing.T) {
		list := &fakeListBreedsUseCase{}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, list, &fakeListBreedDropdownUseCase{}, &fakeUpdateBreedUseCase{}, &fakeDeleteBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/breeds?name=dur&active=true", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		if list.filter.Name == nil || *list.filter.Name != "dur" {
			t.Fatalf("unexpected name filter: %#v", list.filter.Name)
		}
		if list.filter.Active == nil || !*list.filter.Active {
			t.Fatalf("unexpected active filter: %#v", list.filter.Active)
		}
	})

	t.Run("returns bad request for a malformed active filter", func(t *testing.T) {
		list := &fakeListBreedsUseCase{}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, list, &fakeListBreedDropdownUseCase{}, &fakeUpdateBreedUseCase{}, &fakeDeleteBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/breeds?active=maybe", nil))

		if response.Code != http.StatusBadRequest || list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
	})

	t.Run("returns an empty array when there are no breeds", func(t *testing.T) {
		list := &fakeListBreedsUseCase{breeds: []*breeddomain.Breed{}}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, list, &fakeListBreedDropdownUseCase{}, &fakeUpdateBreedUseCase{}, &fakeDeleteBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/breeds", nil))

		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		list := &fakeListBreedsUseCase{err: errors.New("unexpected")}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, list, &fakeListBreedDropdownUseCase{}, &fakeUpdateBreedUseCase{}, &fakeDeleteBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/breeds", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}

func TestBreedHandlerDropdown(t *testing.T) {
	t.Run("returns the breeds with id, name and active", func(t *testing.T) {
		dropdown := &fakeListBreedDropdownUseCase{options: []breeddomain.BreedDropdown{
			{ID: uuid.New(), Name: "Duroc", Active: true},
			{ID: uuid.New(), Name: "Landrace", Active: false},
		}}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, dropdown, &fakeUpdateBreedUseCase{}, &fakeDeleteBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/breeds/dropdown", nil))

		if response.Code != http.StatusOK || !dropdown.called {
			t.Fatalf("status=%d called=%v", response.Code, dropdown.called)
		}
		if dropdown.active {
			t.Fatalf("expected active filter false by default")
		}
		body := strings.TrimSpace(response.Body.String())
		if !strings.HasPrefix(body, "[") || !strings.Contains(body, "Duroc") || !strings.Contains(body, "active") {
			t.Fatalf("unexpected body: %s", body)
		}
		if strings.Contains(body, "created_at") || strings.Contains(body, "created_by") {
			t.Fatalf("dropdown must only include id, name and active: %s", body)
		}
	})

	t.Run("forwards the active query parameter", func(t *testing.T) {
		dropdown := &fakeListBreedDropdownUseCase{}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, dropdown, &fakeUpdateBreedUseCase{}, &fakeDeleteBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/breeds/dropdown?active=true", nil))

		if response.Code != http.StatusOK || !dropdown.active {
			t.Fatalf("status=%d active=%v", response.Code, dropdown.active)
		}
	})

	t.Run("returns bad request for a malformed active value", func(t *testing.T) {
		dropdown := &fakeListBreedDropdownUseCase{}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, dropdown, &fakeUpdateBreedUseCase{}, &fakeDeleteBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/breeds/dropdown?active=maybe", nil))

		if response.Code != http.StatusBadRequest || dropdown.called {
			t.Fatalf("status=%d called=%v", response.Code, dropdown.called)
		}
	})

	t.Run("returns an empty array when there are no breeds", func(t *testing.T) {
		dropdown := &fakeListBreedDropdownUseCase{options: []breeddomain.BreedDropdown{}}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, dropdown, &fakeUpdateBreedUseCase{}, &fakeDeleteBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/breeds/dropdown", nil))

		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		dropdown := &fakeListBreedDropdownUseCase{err: errors.New("unexpected")}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, dropdown, &fakeUpdateBreedUseCase{}, &fakeDeleteBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/breeds/dropdown", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}

func TestBreedHandlerUpdate(t *testing.T) {
	breedID := uuid.New()
	actorID := uuid.New()

	t.Run("updates the provided fields and forwards the actor", func(t *testing.T) {
		update := &fakeUpdateBreedUseCase{breed: handlerBreed()}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, &fakeListBreedDropdownUseCase{}, update, &fakeDeleteBreedUseCase{})
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/breeds/"+breedID.String(), strings.NewReader(`{"name":"New Name","active":false}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || update.breedID != breedID || update.command.UpdatedBy != actorID || update.command.Name == nil || *update.command.Name != "New Name" || update.command.Active == nil || *update.command.Active {
			t.Fatalf("status=%d id=%v command=%#v", response.Code, update.breedID, update.command)
		}
		if body := strings.TrimSpace(response.Body.String()); body != "" {
			t.Fatalf("expected an empty body, got %q", body)
		}
	})

	t.Run("returns bad request for invalid UUID", func(t *testing.T) {
		update := &fakeUpdateBreedUseCase{}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, &fakeListBreedDropdownUseCase{}, update, &fakeDeleteBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPatch, "/api/v1/breeds/not-a-uuid", strings.NewReader(`{"name":"New"}`)))

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		update := &fakeUpdateBreedUseCase{breed: handlerBreed()}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, &fakeListBreedDropdownUseCase{}, update, &fakeDeleteBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPatch, "/api/v1/breeds/"+breedID.String(), strings.NewReader(`{"name":"New"}`)))

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
			{"not found", ports.ErrBreedNotFound, http.StatusNotFound},
			{"duplicate", ports.ErrBreedNameAlreadyUsed, http.StatusConflict},
			{"validation", breeddomain.ErrInvalidUpdate, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				update := &fakeUpdateBreedUseCase{err: test.err}
				handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, &fakeListBreedDropdownUseCase{}, update, &fakeDeleteBreedUseCase{})
				request := httptest.NewRequest(http.MethodPatch, "/api/v1/breeds/"+uuid.New().String(), strings.NewReader(`{"name":"New"}`))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestBreedHandlerDelete(t *testing.T) {
	breedID := uuid.New()

	t.Run("deletes the breed with the provided id", func(t *testing.T) {
		remove := &fakeDeleteBreedUseCase{}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, &fakeListBreedDropdownUseCase{}, &fakeUpdateBreedUseCase{}, remove)
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/breeds/"+breedID.String(), nil)
		request = authenticatedRequest(request, uuid.New())
		response := serve(handler, request)

		if response.Code != http.StatusNoContent || !remove.called || remove.breedID != breedID {
			t.Fatalf("status=%d called=%v id=%v", response.Code, remove.called, remove.breedID)
		}
	})

	t.Run("returns bad request for invalid UUID", func(t *testing.T) {
		remove := &fakeDeleteBreedUseCase{}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, &fakeListBreedDropdownUseCase{}, &fakeUpdateBreedUseCase{}, remove)
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/breeds/not-a-uuid", nil)
		request = authenticatedRequest(request, uuid.New())
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("blocks deletion for non-admins", func(t *testing.T) {
		remove := &fakeDeleteBreedUseCase{}
		handler := newAdminGuardHandler(remove)
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/breeds/"+breedID.String(), nil)
		request = authenticatedRequest(request, uuid.New())
		response := serve(handler, request)

		if response.Code != http.StatusForbidden || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("maps application errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"not found", ports.ErrBreedNotFound, http.StatusNotFound},
			{"in use", ports.ErrBreedInUse, http.StatusConflict},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				remove := &fakeDeleteBreedUseCase{err: test.err}
				handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, &fakeListBreedDropdownUseCase{}, &fakeUpdateBreedUseCase{}, remove)
				request := httptest.NewRequest(http.MethodDelete, "/api/v1/breeds/"+uuid.New().String(), nil)
				request = authenticatedRequest(request, uuid.New())
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}
