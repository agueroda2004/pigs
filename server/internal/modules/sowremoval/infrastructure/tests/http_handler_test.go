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

	sowremovalapplication "server/internal/modules/sowremoval/application"
	sowremovaldomain "server/internal/modules/sowremoval/domain"
	sowremovalinfra "server/internal/modules/sowremoval/infrastructure"
	"server/internal/modules/sowremoval/ports"
)

func TestSowRemovalHandlerCreate(t *testing.T) {
	removal := handlerRemoval()
	actorID := uuid.New()
	sowID := uuid.New()

	body := fmt.Sprintf(
		`{"sow_id":"%s","removal_date":"2026-01-20","type":"Muerte","reason":"Enfermedad","note":"baja"}`,
		sowID,
	)

	t.Run("creates a removal with the actor from context", func(t *testing.T) {
		create := &fakeCreateSowRemovalUseCase{removal: removal}
		handler := newTestHandler(create, &fakeListSowRemovalsUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/sow-removals", strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusCreated || create.command.CreatedBy != actorID {
			t.Fatalf("status=%d command=%#v", response.Code, create.command)
		}
		if create.command.SowID != sowID || create.command.Type != sowremovaldomain.TypeDeath {
			t.Fatalf("unexpected command: %#v", create.command)
		}
		if create.command.Reason != sowremovaldomain.ReasonDisease {
			t.Fatalf("unexpected reason: %#v", create.command)
		}
		if create.command.Note == nil || *create.command.Note != "baja" {
			t.Fatalf("unexpected note: %#v", create.command.Note)
		}
		if !create.command.RemovalDate.Equal(time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected date: %v", create.command.RemovalDate)
		}
		if response.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type: %q", response.Header().Get("Content-Type"))
		}
		if !strings.Contains(response.Body.String(), "Gestando") {
			t.Fatalf("unexpected body: %s", response.Body.String())
		}
	})

	t.Run("rejects an unknown field", func(t *testing.T) {
		create := &fakeCreateSowRemovalUseCase{removal: removal}
		handler := newTestHandler(create, &fakeListSowRemovalsUseCase{})
		invalid := fmt.Sprintf(
			`{"sow_id":"%s","removal_date":"2026-01-20","type":"Muerte","reason":"Enfermedad","last_state":"Gestando"}`,
			sowID,
		)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/sow-removals", strings.NewReader(invalid))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		create := &fakeCreateSowRemovalUseCase{removal: removal}
		handler := newTestHandler(create, &fakeListSowRemovalsUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/sow-removals", strings.NewReader(body)))

		if response.Code != http.StatusUnauthorized || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid JSON", func(t *testing.T) {
		create := &fakeCreateSowRemovalUseCase{removal: removal}
		handler := newTestHandler(create, &fakeListSowRemovalsUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/sow-removals", strings.NewReader(`{"sow_id":`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid fields", func(t *testing.T) {
		create := &fakeCreateSowRemovalUseCase{removal: removal}
		handler := newTestHandler(create, &fakeListSowRemovalsUseCase{})

		for _, invalid := range []string{
			`{"sow_id":"not-a-uuid","removal_date":"2026-01-20","type":"Muerte","reason":"Enfermedad"}`,
			fmt.Sprintf(`{"sow_id":"%s","removal_date":"not-a-date","type":"Muerte","reason":"Enfermedad"}`, sowID),
			fmt.Sprintf(`{"sow_id":"%s","removal_date":"2026-01-20","type":"Mágica","reason":"Enfermedad"}`, sowID),
			fmt.Sprintf(`{"sow_id":"%s","removal_date":"2026-01-20","type":"Muerte","reason":"Mágica"}`, sowID),
		} {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/sow-removals", strings.NewReader(invalid))
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
			{"removal not found", ports.ErrSowRemovalNotFound, http.StatusNotFound},
			{"sow not found", ports.ErrSowNotFound, http.StatusNotFound},
			{"service not found", ports.ErrServiceNotFound, http.StatusNotFound},
			{"abortion not found", ports.ErrAbortionNotFound, http.StatusNotFound},
			{"sow already removed", ports.ErrSowAlreadyRemoved, http.StatusConflict},
			{"sow not removable", sowremovalapplication.ErrSowNotRemovable, http.StatusConflict},
			{"sow not removable domain", sowremovaldomain.ErrSowNotRemovable, http.StatusBadRequest},
			{"date before service", sowremovaldomain.ErrRemovalDateBeforeService, http.StatusBadRequest},
			{"date before abortion", sowremovaldomain.ErrRemovalDateBeforeAbortion, http.StatusBadRequest},
			{"date before entry", sowremovaldomain.ErrRemovalDateBeforeEntry, http.StatusBadRequest},
			{"date in future", sowremovaldomain.ErrRemovalDateInFuture, http.StatusBadRequest},
			{"invalid type", sowremovaldomain.ErrInvalidType, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				create := &fakeCreateSowRemovalUseCase{err: test.err}
				handler := newTestHandler(create, &fakeListSowRemovalsUseCase{})
				request := httptest.NewRequest(http.MethodPost, "/api/v1/sow-removals", strings.NewReader(body))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestSowRemovalHandlerList(t *testing.T) {
	t.Run("returns every removal", func(t *testing.T) {
		list := &fakeListSowRemovalsUseCase{removals: []*sowremovaldomain.SowRemoval{handlerRemoval()}}
		handler := newTestHandler(&fakeCreateSowRemovalUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sow-removals", nil))

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
		list := &fakeListSowRemovalsUseCase{removals: []*sowremovaldomain.SowRemoval{}}
		handler := newTestHandler(&fakeCreateSowRemovalUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sow-removals", nil))

		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("parses the sow_id filter", func(t *testing.T) {
		sowID := uuid.New()
		list := &fakeListSowRemovalsUseCase{}
		handler := newTestHandler(&fakeCreateSowRemovalUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sow-removals?sow_id="+sowID.String(), nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		if list.filter.SowID == nil || *list.filter.SowID != sowID {
			t.Fatalf("unexpected sow id: %#v", list.filter.SowID)
		}
	})

	t.Run("returns an empty filter when no params are given", func(t *testing.T) {
		list := &fakeListSowRemovalsUseCase{}
		handler := newTestHandler(&fakeCreateSowRemovalUseCase{}, list)
		serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sow-removals", nil))

		if list.filter.SowID != nil {
			t.Fatalf("unexpected filter: %#v", list.filter)
		}
	})

	t.Run("rejects invalid filter values", func(t *testing.T) {
		list := &fakeListSowRemovalsUseCase{}
		handler := newTestHandler(&fakeCreateSowRemovalUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sow-removals?sow_id=not-a-uuid", nil))

		if response.Code != http.StatusBadRequest || list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		list := &fakeListSowRemovalsUseCase{err: errors.New("unexpected")}
		handler := newTestHandler(&fakeCreateSowRemovalUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sow-removals", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}

func TestSowRemovalHandlerUpdate(t *testing.T) {
	removal := handlerRemoval()
	removalID := removal.ID
	actorID := uuid.New()

	newUpdateHandler := func(update *fakeUpdateSowRemovalUseCase) *sowremovalinfra.SowRemovalHandler {
		return newTestHandlerWithUpdate(&fakeCreateSowRemovalUseCase{}, update, &fakeListSowRemovalsUseCase{})
	}

	t.Run("updates the removal with the actor from context", func(t *testing.T) {
		update := &fakeUpdateSowRemovalUseCase{removal: removal}
		handler := newUpdateHandler(update)
		body := `{"removal_date":"2026-01-25","type":"Desecho","reason":"Otro","note":"cambiada"}`
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/sow-removals/"+removalID.String(), strings.NewReader(body))
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
		if update.command.Type == nil || *update.command.Type != sowremovaldomain.TypeDiscard {
			t.Fatalf("unexpected type: %#v", update.command.Type)
		}
		if update.command.Reason == nil || *update.command.Reason != sowremovaldomain.ReasonOther {
			t.Fatalf("unexpected reason: %#v", update.command.Reason)
		}
		if update.command.Note == nil || *update.command.Note != "cambiada" {
			t.Fatalf("unexpected note: %#v", update.command.Note)
		}
	})

	t.Run("leaves omitted fields nil", func(t *testing.T) {
		update := &fakeUpdateSowRemovalUseCase{removal: removal}
		handler := newUpdateHandler(update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/sow-removals/"+removalID.String(), strings.NewReader(`{"reason":"Otro"}`))
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
		update := &fakeUpdateSowRemovalUseCase{removal: removal}
		handler := newUpdateHandler(update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/sow-removals/not-a-uuid", strings.NewReader(`{"reason":"Otro"}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("rejects invalid fields", func(t *testing.T) {
		update := &fakeUpdateSowRemovalUseCase{removal: removal}
		handler := newUpdateHandler(update)

		for _, invalid := range []string{
			`{"removal_date":"not-a-date"}`,
			`{"type":"Mágica"}`,
			`{"reason":"Mágica"}`,
		} {
			request := httptest.NewRequest(http.MethodPatch, "/api/v1/sow-removals/"+removalID.String(), strings.NewReader(invalid))
			request = authenticatedRequest(request, actorID)
			response := serve(handler, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("body=%s status=%d", invalid, response.Code)
			}
		}
	})

	t.Run("rejects the immutable sow field", func(t *testing.T) {
		update := &fakeUpdateSowRemovalUseCase{removal: removal}
		handler := newUpdateHandler(update)
		body := fmt.Sprintf(`{"sow_id":"%s","reason":"Otro"}`, uuid.New())
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/sow-removals/"+removalID.String(), strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		update := &fakeUpdateSowRemovalUseCase{removal: removal}
		handler := newUpdateHandler(update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/sow-removals/"+removalID.String(), strings.NewReader(`{"reason":"Otro"}`))
		response := serve(handler, request)

		if response.Code != http.StatusUnauthorized || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("maps not found", func(t *testing.T) {
		update := &fakeUpdateSowRemovalUseCase{err: ports.ErrSowRemovalNotFound}
		handler := newUpdateHandler(update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/sow-removals/"+removalID.String(), strings.NewReader(`{"reason":"Otro"}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusNotFound {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusNotFound)
		}
	})

	t.Run("maps a date before entry to bad request", func(t *testing.T) {
		update := &fakeUpdateSowRemovalUseCase{err: sowremovaldomain.ErrRemovalDateBeforeEntry}
		handler := newUpdateHandler(update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/sow-removals/"+removalID.String(), strings.NewReader(`{"removal_date":"2025-01-01"}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusBadRequest)
		}
	})
}

func TestSowRemovalHandlerDelete(t *testing.T) {
	removalID := uuid.New()
	actorID := uuid.New()

	newDeleteHandler := func(deleteUseCase *fakeDeleteSowRemovalUseCase) *sowremovalinfra.SowRemovalHandler {
		return newTestHandlerWithDelete(
			&fakeCreateSowRemovalUseCase{},
			&fakeUpdateSowRemovalUseCase{},
			deleteUseCase,
			&fakeListSowRemovalsUseCase{},
		)
	}

	t.Run("deletes the removal with the actor from context", func(t *testing.T) {
		deleteUseCase := &fakeDeleteSowRemovalUseCase{}
		handler := newDeleteHandler(deleteUseCase)
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/sow-removals/"+removalID.String(), nil)
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
		deleteUseCase := &fakeDeleteSowRemovalUseCase{}
		handler := newDeleteHandler(deleteUseCase)
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/sow-removals/not-a-uuid", nil)
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || deleteUseCase.called {
			t.Fatalf("status=%d called=%v", response.Code, deleteUseCase.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		deleteUseCase := &fakeDeleteSowRemovalUseCase{}
		handler := newDeleteHandler(deleteUseCase)
		response := serve(handler, httptest.NewRequest(http.MethodDelete, "/api/v1/sow-removals/"+removalID.String(), nil))

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
			{"removal not found", ports.ErrSowRemovalNotFound, http.StatusNotFound},
			{"sow not found", ports.ErrSowNotFound, http.StatusNotFound},
			{"state mismatch", sowremovalapplication.ErrSowStateMismatch, http.StatusConflict},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				deleteUseCase := &fakeDeleteSowRemovalUseCase{err: test.err}
				handler := newDeleteHandler(deleteUseCase)
				request := httptest.NewRequest(http.MethodDelete, "/api/v1/sow-removals/"+removalID.String(), nil)
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)

				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}
