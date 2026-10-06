package tests

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	boarremovalapplication "server/internal/modules/boarremoval/application"
	boarremovaldomain "server/internal/modules/boarremoval/domain"
	boarremovalinfra "server/internal/modules/boarremoval/infrastructure"
	"server/internal/modules/boarremoval/ports"
)

func TestBoarRemovalHandlerCreate(t *testing.T) {
	removal := handlerRemoval()
	actorID := uuid.New()
	boarID := uuid.New()

	body := fmt.Sprintf(
		`{"boar_id":"%s","removal_date":"2026-01-20","type":"Muerte","reason":"Enfermedad","note":"baja"}`,
		boarID,
	)

	t.Run("creates a removal with the actor from context", func(t *testing.T) {
		create := &fakeCreateBoarRemovalUseCase{removal: removal}
		handler := newTestHandler(create, &fakeListBoarRemovalsUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/boar-removals", strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusCreated || create.command.CreatedBy != actorID {
			t.Fatalf("status=%d command=%#v", response.Code, create.command)
		}
		if create.command.BoarID != boarID || create.command.Type != boarremovaldomain.TypeDeath {
			t.Fatalf("unexpected command: %#v", create.command)
		}
		if create.command.Reason != boarremovaldomain.ReasonDisease {
			t.Fatalf("unexpected reason: %#v", create.command)
		}
		if create.command.Note == nil || *create.command.Note != "baja" {
			t.Fatalf("unexpected note: %#v", create.command.Note)
		}
		if !create.command.RemovalDate.Equal(time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected date: %v", create.command.RemovalDate)
		}
		if body := strings.TrimSpace(response.Body.String()); body != "" {
			t.Fatalf("expected an empty body, got %q", body)
		}
	})

	t.Run("rejects an unknown field", func(t *testing.T) {
		create := &fakeCreateBoarRemovalUseCase{removal: removal}
		handler := newTestHandler(create, &fakeListBoarRemovalsUseCase{})
		invalid := fmt.Sprintf(
			`{"boar_id":"%s","removal_date":"2026-01-20","type":"Muerte","reason":"Enfermedad","last_state":"Vivo"}`,
			boarID,
		)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/boar-removals", strings.NewReader(invalid))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		create := &fakeCreateBoarRemovalUseCase{removal: removal}
		handler := newTestHandler(create, &fakeListBoarRemovalsUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/boar-removals", strings.NewReader(body)))

		if response.Code != http.StatusUnauthorized || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid JSON", func(t *testing.T) {
		create := &fakeCreateBoarRemovalUseCase{removal: removal}
		handler := newTestHandler(create, &fakeListBoarRemovalsUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/boar-removals", strings.NewReader(`{"boar_id":`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid fields", func(t *testing.T) {
		create := &fakeCreateBoarRemovalUseCase{removal: removal}
		handler := newTestHandler(create, &fakeListBoarRemovalsUseCase{})

		for _, invalid := range []string{
			`{"boar_id":"not-a-uuid","removal_date":"2026-01-20","type":"Muerte","reason":"Enfermedad"}`,
			fmt.Sprintf(`{"boar_id":"%s","removal_date":"not-a-date","type":"Muerte","reason":"Enfermedad"}`, boarID),
			fmt.Sprintf(`{"boar_id":"%s","removal_date":"2026-01-20","type":"Mágica","reason":"Enfermedad"}`, boarID),
			fmt.Sprintf(`{"boar_id":"%s","removal_date":"2026-01-20","type":"Muerte","reason":"Mágica"}`, boarID),
		} {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/boar-removals", strings.NewReader(invalid))
			request = authenticatedRequest(request, actorID)
			response := serve(handler, request)
			if response.Code != http.StatusBadRequest || create.called {
				t.Fatalf("status=%d called=%v body=%s", response.Code, create.called, invalid)
			}
		}
	})

	t.Run("maps application and domain errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"removal not found", ports.ErrBoarRemovalNotFound, http.StatusNotFound},
			{"boar not found", ports.ErrBoarNotFound, http.StatusNotFound},
			{"boar already removed", ports.ErrBoarAlreadyRemoved, http.StatusConflict},
			{"boar not removable", boarremovalapplication.ErrBoarNotRemovable, http.StatusConflict},
			{"boar not removable domain", boarremovaldomain.ErrBoarNotRemovable, http.StatusBadRequest},
			{"date before entry", boarremovaldomain.ErrRemovalDateBeforeEntry, http.StatusBadRequest},
			{"date before mount", boarremovaldomain.ErrRemovalDateBeforeMount, http.StatusBadRequest},
			{"date in future", boarremovaldomain.ErrRemovalDateInFuture, http.StatusBadRequest},
			{"invalid type", boarremovaldomain.ErrInvalidType, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				create := &fakeCreateBoarRemovalUseCase{err: test.err}
				handler := newTestHandler(create, &fakeListBoarRemovalsUseCase{})
				request := httptest.NewRequest(http.MethodPost, "/api/v1/boar-removals", strings.NewReader(body))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestBoarRemovalHandlerList(t *testing.T) {
	t.Run("returns every removal", func(t *testing.T) {
		list := &fakeListBoarRemovalsUseCase{removals: []*boarremovaldomain.BoarRemoval{handlerRemoval()}}
		handler := newTestHandler(&fakeCreateBoarRemovalUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/boar-removals", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		body := response.Body.String()
		for _, want := range []string{"Muerte", "Enfermedad", "last_state", "2026-01-20"} {
			if !strings.Contains(body, want) {
				t.Fatalf("body missing %q: %s", want, body)
			}
		}
	})

	t.Run("returns an empty array when there are no removals", func(t *testing.T) {
		list := &fakeListBoarRemovalsUseCase{removals: []*boarremovaldomain.BoarRemoval{}}
		handler := newTestHandler(&fakeCreateBoarRemovalUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/boar-removals", nil))

		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("parses the boar_id filter", func(t *testing.T) {
		boarID := uuid.New()
		list := &fakeListBoarRemovalsUseCase{}
		handler := newTestHandler(&fakeCreateBoarRemovalUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/boar-removals?boar_id="+boarID.String(), nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		if list.filter.BoarID == nil || *list.filter.BoarID != boarID {
			t.Fatalf("unexpected boar id: %#v", list.filter.BoarID)
		}
	})

	t.Run("returns an empty filter when no params are given", func(t *testing.T) {
		list := &fakeListBoarRemovalsUseCase{}
		handler := newTestHandler(&fakeCreateBoarRemovalUseCase{}, list)
		serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/boar-removals", nil))

		if list.filter.BoarID != nil {
			t.Fatalf("unexpected filter: %#v", list.filter)
		}
	})

	t.Run("rejects invalid filter values", func(t *testing.T) {
		list := &fakeListBoarRemovalsUseCase{}
		handler := newTestHandler(&fakeCreateBoarRemovalUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/boar-removals?boar_id=not-a-uuid", nil))

		if response.Code != http.StatusBadRequest || list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		list := &fakeListBoarRemovalsUseCase{err: errors.New("unexpected")}
		handler := newTestHandler(&fakeCreateBoarRemovalUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/boar-removals", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}

func TestBoarRemovalHandlerUpdate(t *testing.T) {
	removal := handlerRemoval()
	removalID := removal.ID
	actorID := uuid.New()

	newUpdateHandler := func(update *fakeUpdateBoarRemovalUseCase) *boarremovalinfra.BoarRemovalHandler {
		return newTestHandlerWithUpdate(&fakeCreateBoarRemovalUseCase{}, update, &fakeListBoarRemovalsUseCase{})
	}

	t.Run("updates the removal with the actor from context", func(t *testing.T) {
		update := &fakeUpdateBoarRemovalUseCase{removal: removal}
		handler := newUpdateHandler(update)
		body := `{"removal_date":"2026-01-25","type":"Desecho","reason":"Otro","note":"cambiada"}`
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/boar-removals/"+removalID.String(), strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || !update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
		if update.id != removalID || update.command.UpdatedBy != actorID {
			t.Fatalf("unexpected command: %#v", update.command)
		}
		if update.command.RemovalDate == nil || !update.command.RemovalDate.Equal(time.Date(2026, time.January, 25, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected date: %#v", update.command.RemovalDate)
		}
		if update.command.Type == nil || *update.command.Type != boarremovaldomain.TypeDiscard {
			t.Fatalf("unexpected type: %#v", update.command.Type)
		}
		if update.command.Reason == nil || *update.command.Reason != boarremovaldomain.ReasonOther {
			t.Fatalf("unexpected reason: %#v", update.command.Reason)
		}
		if update.command.Note == nil || *update.command.Note != "cambiada" {
			t.Fatalf("unexpected note: %#v", update.command.Note)
		}
	})

	t.Run("leaves omitted fields nil", func(t *testing.T) {
		update := &fakeUpdateBoarRemovalUseCase{removal: removal}
		handler := newUpdateHandler(update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/boar-removals/"+removalID.String(), strings.NewReader(`{"reason":"Otro"}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK {
			t.Fatalf("status=%d", response.Code)
		}
		if update.command.RemovalDate != nil || update.command.Type != nil || update.command.Note != nil {
			t.Fatalf("expected omitted fields to stay nil: %#v", update.command)
		}
	})

	t.Run("rejects an invalid identifier", func(t *testing.T) {
		update := &fakeUpdateBoarRemovalUseCase{removal: removal}
		handler := newUpdateHandler(update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/boar-removals/not-a-uuid", strings.NewReader(`{"reason":"Otro"}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("rejects invalid fields", func(t *testing.T) {
		update := &fakeUpdateBoarRemovalUseCase{removal: removal}
		handler := newUpdateHandler(update)

		for _, invalid := range []string{
			`{"removal_date":"not-a-date"}`,
			`{"type":"Mágica"}`,
			`{"reason":"Mágica"}`,
		} {
			request := httptest.NewRequest(http.MethodPatch, "/api/v1/boar-removals/"+removalID.String(), strings.NewReader(invalid))
			request = authenticatedRequest(request, actorID)
			response := serve(handler, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("body=%s status=%d", invalid, response.Code)
			}
		}
	})

	t.Run("rejects the immutable boar field", func(t *testing.T) {
		update := &fakeUpdateBoarRemovalUseCase{removal: removal}
		handler := newUpdateHandler(update)
		body := fmt.Sprintf(`{"boar_id":"%s","reason":"Otro"}`, uuid.New())
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/boar-removals/"+removalID.String(), strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		update := &fakeUpdateBoarRemovalUseCase{removal: removal}
		handler := newUpdateHandler(update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/boar-removals/"+removalID.String(), strings.NewReader(`{"reason":"Otro"}`))
		response := serve(handler, request)

		if response.Code != http.StatusUnauthorized || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("maps application and domain errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"removal not found", ports.ErrBoarRemovalNotFound, http.StatusNotFound},
			{"date before entry", boarremovaldomain.ErrRemovalDateBeforeEntry, http.StatusBadRequest},
			{"date before mount", boarremovaldomain.ErrRemovalDateBeforeMount, http.StatusBadRequest},
			{"invalid update", boarremovaldomain.ErrInvalidUpdate, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				update := &fakeUpdateBoarRemovalUseCase{err: test.err}
				handler := newUpdateHandler(update)
				request := httptest.NewRequest(http.MethodPatch, "/api/v1/boar-removals/"+removalID.String(), strings.NewReader(`{"reason":"Otro"}`))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)

				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestBoarRemovalHandlerDelete(t *testing.T) {
	removalID := uuid.New()
	actorID := uuid.New()

	newDeleteHandler := func(deleteUseCase *fakeDeleteBoarRemovalUseCase) *boarremovalinfra.BoarRemovalHandler {
		return newTestHandlerWithDelete(
			&fakeCreateBoarRemovalUseCase{},
			&fakeUpdateBoarRemovalUseCase{},
			deleteUseCase,
			&fakeListBoarRemovalsUseCase{},
		)
	}

	t.Run("deletes the removal with the actor from context", func(t *testing.T) {
		deleteUseCase := &fakeDeleteBoarRemovalUseCase{}
		handler := newDeleteHandler(deleteUseCase)
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/boar-removals/"+removalID.String(), nil)
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusNoContent || !deleteUseCase.called {
			t.Fatalf("status=%d called=%v", response.Code, deleteUseCase.called)
		}
		if deleteUseCase.command.ID != removalID || deleteUseCase.command.DeletedBy != actorID {
			t.Fatalf("unexpected command: %#v", deleteUseCase.command)
		}
		if response.Body.Len() != 0 {
			t.Fatalf("expected empty body, got %s", response.Body.String())
		}
	})

	t.Run("rejects an invalid identifier", func(t *testing.T) {
		deleteUseCase := &fakeDeleteBoarRemovalUseCase{}
		handler := newDeleteHandler(deleteUseCase)
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/boar-removals/not-a-uuid", nil)
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || deleteUseCase.called {
			t.Fatalf("status=%d called=%v", response.Code, deleteUseCase.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		deleteUseCase := &fakeDeleteBoarRemovalUseCase{}
		handler := newDeleteHandler(deleteUseCase)
		response := serve(handler, httptest.NewRequest(http.MethodDelete, "/api/v1/boar-removals/"+removalID.String(), nil))

		if response.Code != http.StatusUnauthorized || deleteUseCase.called {
			t.Fatalf("status=%d called=%v", response.Code, deleteUseCase.called)
		}
	})

	t.Run("maps application and port errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"removal not found", ports.ErrBoarRemovalNotFound, http.StatusNotFound},
			{"boar not found", ports.ErrBoarNotFound, http.StatusNotFound},
			{"state mismatch", boarremovalapplication.ErrBoarStateMismatch, http.StatusConflict},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				deleteUseCase := &fakeDeleteBoarRemovalUseCase{err: test.err}
				handler := newDeleteHandler(deleteUseCase)
				request := httptest.NewRequest(http.MethodDelete, "/api/v1/boar-removals/"+removalID.String(), nil)
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)

				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}
