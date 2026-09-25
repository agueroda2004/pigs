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
		handler := newTestHandler(create, &fakeListBreedsUseCase{}, &fakeListBreedOptionsUseCase{}, &fakeUpdateBreedUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/breeds", strings.NewReader(`{"name":"Duroc"}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusCreated || create.command.CreatedBy != actorID || create.command.Name != "Duroc" {
			t.Fatalf("status=%d command=%#v", response.Code, create.command)
		}
		if response.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type: %q", response.Header().Get("Content-Type"))
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		create := &fakeCreateBreedUseCase{breed: breed}
		handler := newTestHandler(create, &fakeListBreedsUseCase{}, &fakeListBreedOptionsUseCase{}, &fakeUpdateBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/breeds", strings.NewReader(`{"name":"Duroc"}`)))

		if response.Code != http.StatusUnauthorized || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid JSON", func(t *testing.T) {
		create := &fakeCreateBreedUseCase{breed: breed}
		handler := newTestHandler(create, &fakeListBreedsUseCase{}, &fakeListBreedOptionsUseCase{}, &fakeUpdateBreedUseCase{})
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
				handler := newTestHandler(create, &fakeListBreedsUseCase{}, &fakeListBreedOptionsUseCase{}, &fakeUpdateBreedUseCase{})
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
	t.Run("returns every breed", func(t *testing.T) {
		list := &fakeListBreedsUseCase{breeds: []*breeddomain.Breed{handlerBreed(), handlerBreed()}}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, list, &fakeListBreedOptionsUseCase{}, &fakeUpdateBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/breeds", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		body := strings.TrimSpace(response.Body.String())
		if !strings.HasPrefix(body, "[") || !strings.Contains(body, "Duroc") {
			t.Fatalf("unexpected body: %s", body)
		}
	})

	t.Run("returns an empty array when there are no breeds", func(t *testing.T) {
		list := &fakeListBreedsUseCase{breeds: []*breeddomain.Breed{}}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, list, &fakeListBreedOptionsUseCase{}, &fakeUpdateBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/breeds", nil))

		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		list := &fakeListBreedsUseCase{err: errors.New("unexpected")}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, list, &fakeListBreedOptionsUseCase{}, &fakeUpdateBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/breeds", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}

func TestBreedHandlerListOptions(t *testing.T) {
	t.Run("returns the active breed options with id and name", func(t *testing.T) {
		options := &fakeListBreedOptionsUseCase{options: []breeddomain.BreedOption{
			{ID: uuid.New(), Name: "Duroc"},
			{ID: uuid.New(), Name: "Landrace"},
		}}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, options, &fakeUpdateBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/breeds/options", nil))

		if response.Code != http.StatusOK || !options.called {
			t.Fatalf("status=%d called=%v", response.Code, options.called)
		}
		body := strings.TrimSpace(response.Body.String())
		if !strings.HasPrefix(body, "[") || !strings.Contains(body, "Duroc") || !strings.Contains(body, "Landrace") {
			t.Fatalf("unexpected body: %s", body)
		}
		if strings.Contains(body, "active") || strings.Contains(body, "created_at") {
			t.Fatalf("options must only include id and name: %s", body)
		}
	})

	t.Run("returns an empty array when there are no active breeds", func(t *testing.T) {
		options := &fakeListBreedOptionsUseCase{options: []breeddomain.BreedOption{}}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, options, &fakeUpdateBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/breeds/options", nil))

		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		options := &fakeListBreedOptionsUseCase{err: errors.New("unexpected")}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, options, &fakeUpdateBreedUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/breeds/options", nil))

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
		handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, &fakeListBreedOptionsUseCase{}, update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/breeds/"+breedID.String(), strings.NewReader(`{"name":"New Name","active":false}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || update.breedID != breedID || update.command.UpdatedBy != actorID || update.command.Name == nil || *update.command.Name != "New Name" || update.command.Active == nil || *update.command.Active {
			t.Fatalf("status=%d id=%v command=%#v", response.Code, update.breedID, update.command)
		}
	})

	t.Run("returns bad request for invalid UUID", func(t *testing.T) {
		update := &fakeUpdateBreedUseCase{}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, &fakeListBreedOptionsUseCase{}, update)
		response := serve(handler, httptest.NewRequest(http.MethodPatch, "/api/v1/breeds/not-a-uuid", strings.NewReader(`{"name":"New"}`)))

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		update := &fakeUpdateBreedUseCase{breed: handlerBreed()}
		handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, &fakeListBreedOptionsUseCase{}, update)
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
				handler := newTestHandler(&fakeCreateBreedUseCase{}, &fakeListBreedsUseCase{}, &fakeListBreedOptionsUseCase{}, update)
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
