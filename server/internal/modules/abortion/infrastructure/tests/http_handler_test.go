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

	abortionapplication "server/internal/modules/abortion/application"
	abortiondomain "server/internal/modules/abortion/domain"
	"server/internal/modules/abortion/ports"
)

func TestAbortionHandlerCreate(t *testing.T) {
	abortion := handlerAbortion()
	actorID := uuid.New()
	sowID := uuid.New()

	body := fmt.Sprintf(
		`{"sow_id":"%s","abortion_date":"2026-01-15","cause":"Infeccioso","note":"aborto"}`,
		sowID,
	)

	t.Run("creates an abortion with the actor from context", func(t *testing.T) {
		create := &fakeCreateAbortionUseCase{abortion: abortion}
		handler := newTestHandler(create, &fakeListAbortionsUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/abortions", strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusCreated || create.command.CreatedBy != actorID {
			t.Fatalf("status=%d command=%#v", response.Code, create.command)
		}
		if create.command.SowID != sowID || create.command.Cause != abortiondomain.CauseInfectious {
			t.Fatalf("unexpected command: %#v", create.command)
		}
		if create.command.Note == nil || *create.command.Note != "aborto" {
			t.Fatalf("unexpected note: %#v", create.command.Note)
		}
		if !create.command.AbortionDate.Equal(time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected date: %v", create.command.AbortionDate)
		}
		if strings.TrimSpace(response.Body.String()) != "" {
			t.Fatalf("expected an empty body, got: %s", response.Body.String())
		}
	})

	t.Run("rejects an unknown field", func(t *testing.T) {
		create := &fakeCreateAbortionUseCase{abortion: abortion}
		handler := newTestHandler(create, &fakeListAbortionsUseCase{})
		invalid := fmt.Sprintf(
			`{"sow_id":"%s","abortion_date":"2026-01-15","cause":"Infeccioso","service_id":"%s"}`,
			sowID, uuid.New(),
		)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/abortions", strings.NewReader(invalid))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		create := &fakeCreateAbortionUseCase{abortion: abortion}
		handler := newTestHandler(create, &fakeListAbortionsUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/abortions", strings.NewReader(body)))

		if response.Code != http.StatusUnauthorized || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid JSON", func(t *testing.T) {
		create := &fakeCreateAbortionUseCase{abortion: abortion}
		handler := newTestHandler(create, &fakeListAbortionsUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/abortions", strings.NewReader(`{"sow_id":`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid fields", func(t *testing.T) {
		create := &fakeCreateAbortionUseCase{abortion: abortion}
		handler := newTestHandler(create, &fakeListAbortionsUseCase{})

		for _, invalid := range []string{
			`{"sow_id":"not-a-uuid","abortion_date":"2026-01-15","cause":"Infeccioso"}`,
			fmt.Sprintf(`{"sow_id":"%s","abortion_date":"not-a-date","cause":"Infeccioso"}`, sowID),
			fmt.Sprintf(`{"sow_id":"%s","abortion_date":"2026-01-15","cause":"Mágica"}`, sowID),
		} {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/abortions", strings.NewReader(invalid))
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
			{"abortion not found", ports.ErrAbortionNotFound, http.StatusNotFound},
			{"sow not found", ports.ErrSowNotFound, http.StatusNotFound},
			{"service not found", ports.ErrServiceNotFound, http.StatusNotFound},
			{"sow not gestating", abortionapplication.ErrSowNotGestating, http.StatusConflict},
			{"service not confirmed", abortionapplication.ErrServiceNotConfirmed, http.StatusConflict},
			{"date before mount", abortiondomain.ErrAbortionDateBeforeMount, http.StatusBadRequest},
			{"date in future", abortiondomain.ErrAbortionDateInFuture, http.StatusBadRequest},
			{"invalid cause", abortiondomain.ErrInvalidCause, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				create := &fakeCreateAbortionUseCase{err: test.err}
				handler := newTestHandler(create, &fakeListAbortionsUseCase{})
				request := httptest.NewRequest(http.MethodPost, "/api/v1/abortions", strings.NewReader(body))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestAbortionHandlerList(t *testing.T) {
	t.Run("returns every abortion", func(t *testing.T) {
		list := &fakeListAbortionsUseCase{abortions: []*abortiondomain.Abortion{handlerAbortion()}}
		handler := newTestHandler(&fakeCreateAbortionUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/abortions", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		body := response.Body.String()
		for _, want := range []string{"Infeccioso", "2026-01-15", "service_id", "sow_code", "C-001"} {
			if !strings.Contains(body, want) {
				t.Fatalf("body missing %q: %s", want, body)
			}
		}
		for _, unwanted := range []string{"created_at", "updated_at", "created_by", "updated_by"} {
			if strings.Contains(body, unwanted) {
				t.Fatalf("list must omit audit fields, found %q: %s", unwanted, body)
			}
		}
	})

	t.Run("returns an empty items array when there are no abortions", func(t *testing.T) {
		list := &fakeListAbortionsUseCase{abortions: []*abortiondomain.Abortion{}}
		handler := newTestHandler(&fakeCreateAbortionUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/abortions", nil))

		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"items":[]`) {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("parses the page query parameter", func(t *testing.T) {
		list := &fakeListAbortionsUseCase{}
		handler := newTestHandler(&fakeCreateAbortionUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/abortions?page=3", nil))

		if response.Code != http.StatusOK || list.page != 3 {
			t.Fatalf("status=%d page=%d", response.Code, list.page)
		}
	})

	t.Run("rejects an invalid page", func(t *testing.T) {
		list := &fakeListAbortionsUseCase{}
		handler := newTestHandler(&fakeCreateAbortionUseCase{}, list)

		for _, target := range []string{"/api/v1/abortions?page=0", "/api/v1/abortions?page=abc"} {
			response := serve(handler, httptest.NewRequest(http.MethodGet, target, nil))
			if response.Code != http.StatusBadRequest || list.called {
				t.Fatalf("target=%s status=%d called=%v", target, response.Code, list.called)
			}
		}
	})

	t.Run("parses the sow_code filter", func(t *testing.T) {
		list := &fakeListAbortionsUseCase{}
		handler := newTestHandler(&fakeCreateAbortionUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/abortions?sow_code=C-001", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		if list.filter.SowCode == nil || *list.filter.SowCode != "C-001" {
			t.Fatalf("unexpected sow code: %#v", list.filter.SowCode)
		}
	})

	t.Run("trims the sow code filter", func(t *testing.T) {
		list := &fakeListAbortionsUseCase{}
		handler := newTestHandler(&fakeCreateAbortionUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/abortions?sow_code=%20C-001%20", nil))

		if response.Code != http.StatusOK || list.filter.SowCode == nil || *list.filter.SowCode != "C-001" {
			t.Fatalf("status=%d filter=%#v", response.Code, list.filter.SowCode)
		}
	})

	t.Run("returns an empty filter when no params are given", func(t *testing.T) {
		list := &fakeListAbortionsUseCase{}
		handler := newTestHandler(&fakeCreateAbortionUseCase{}, list)
		serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/abortions", nil))

		if list.filter.SowCode != nil {
			t.Fatalf("unexpected filter: %#v", list.filter)
		}
	})

	t.Run("rejects invalid filter values", func(t *testing.T) {
		list := &fakeListAbortionsUseCase{}
		handler := newTestHandler(&fakeCreateAbortionUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/abortions?sow_code="+strings.Repeat("A", 51), nil))

		if response.Code != http.StatusBadRequest || list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		list := &fakeListAbortionsUseCase{err: errors.New("unexpected")}
		handler := newTestHandler(&fakeCreateAbortionUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/abortions", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}

func TestAbortionHandlerUpdate(t *testing.T) {
	abortionID := uuid.New()
	actorID := uuid.New()

	t.Run("applies the update and returns no body", func(t *testing.T) {
		update := &fakeUpdateAbortionUseCase{}
		handler := newTestHandlerWithUpdate(&fakeCreateAbortionUseCase{}, &fakeListAbortionsUseCase{}, update)
		body := `{"abortion_date":"2026-01-20","cause":"Traumatismo","note":"actualizado"}`
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/abortions/"+abortionID.String(), strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || !update.called || update.id != abortionID {
			t.Fatalf("status=%d id=%v command=%#v", response.Code, update.id, update.command)
		}
		if update.command.UpdatedBy != actorID {
			t.Fatalf("unexpected actor: %#v", update.command)
		}
		if update.command.AbortionDate == nil || update.command.Cause == nil || update.command.Note == nil {
			t.Fatalf("unexpected command: %#v", update.command)
		}
		if !update.command.AbortionDate.Equal(time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected date: %v", update.command.AbortionDate)
		}
		if *update.command.Cause != abortiondomain.CauseTrauma || *update.command.Note != "actualizado" {
			t.Fatalf("unexpected command: %#v", update.command)
		}
		if strings.TrimSpace(response.Body.String()) != "" {
			t.Fatalf("expected an empty body, got: %s", response.Body.String())
		}
	})

	t.Run("allows updating only the cause", func(t *testing.T) {
		update := &fakeUpdateAbortionUseCase{}
		handler := newTestHandlerWithUpdate(&fakeCreateAbortionUseCase{}, &fakeListAbortionsUseCase{}, update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/abortions/"+abortionID.String(), strings.NewReader(`{"cause":"Otro"}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || !update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
		if update.command.AbortionDate != nil || update.command.Note != nil {
			t.Fatalf("unexpected command: %#v", update.command)
		}
	})

	t.Run("rejects an invalid id and fields", func(t *testing.T) {
		update := &fakeUpdateAbortionUseCase{}
		handler := newTestHandlerWithUpdate(&fakeCreateAbortionUseCase{}, &fakeListAbortionsUseCase{}, update)

		invalidID := httptest.NewRequest(http.MethodPatch, "/api/v1/abortions/not-a-uuid", strings.NewReader(`{"cause":"Otro"}`))
		invalidID = authenticatedRequest(invalidID, actorID)
		if response := serve(handler, invalidID); response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("invalid id: status=%d called=%v", response.Code, update.called)
		}

		for _, body := range []string{
			`{"abortion_date":"not-a-date"}`,
			`{"cause":"Mágica"}`,
			`{"sow_id":"` + uuid.New().String() + `"}`,
		} {
			request := httptest.NewRequest(http.MethodPatch, "/api/v1/abortions/"+abortionID.String(), strings.NewReader(body))
			request = authenticatedRequest(request, actorID)
			if response := serve(handler, request); response.Code != http.StatusBadRequest || update.called {
				t.Fatalf("body=%s status=%d called=%v", body, response.Code, update.called)
			}
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		update := &fakeUpdateAbortionUseCase{}
		handler := newTestHandlerWithUpdate(&fakeCreateAbortionUseCase{}, &fakeListAbortionsUseCase{}, update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/abortions/"+abortionID.String(), strings.NewReader(`{"cause":"Otro"}`))
		response := serve(handler, request)

		if response.Code != http.StatusUnauthorized || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("maps errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"not found", ports.ErrAbortionNotFound, http.StatusNotFound},
			{"before mount", abortiondomain.ErrAbortionDateBeforeMount, http.StatusBadRequest},
			{"before entry", abortiondomain.ErrAbortionDateBeforeEntry, http.StatusBadRequest},
			{"in future", abortiondomain.ErrAbortionDateInFuture, http.StatusBadRequest},
			{"no fields", abortiondomain.ErrInvalidUpdate, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				update := &fakeUpdateAbortionUseCase{err: test.err}
				handler := newTestHandlerWithUpdate(&fakeCreateAbortionUseCase{}, &fakeListAbortionsUseCase{}, update)
				request := httptest.NewRequest(http.MethodPatch, "/api/v1/abortions/"+uuid.New().String(), strings.NewReader(`{"cause":"Otro"}`))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestAbortionHandlerDelete(t *testing.T) {
	abortionID := uuid.New()
	actorID := uuid.New()

	t.Run("deletes the abortion and returns no body", func(t *testing.T) {
		remove := &fakeDeleteAbortionUseCase{}
		handler := newTestHandlerWithDelete(&fakeCreateAbortionUseCase{}, &fakeListAbortionsUseCase{}, remove)
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/abortions/"+abortionID.String(), nil)
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusNoContent || !remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
		if remove.command.ID != abortionID || remove.command.DeletedBy != actorID {
			t.Fatalf("unexpected command: %#v", remove.command)
		}
		if strings.TrimSpace(response.Body.String()) != "" {
			t.Fatalf("expected an empty body, got: %s", response.Body.String())
		}
	})

	t.Run("returns bad request for an invalid id", func(t *testing.T) {
		remove := &fakeDeleteAbortionUseCase{}
		handler := newTestHandlerWithDelete(&fakeCreateAbortionUseCase{}, &fakeListAbortionsUseCase{}, remove)
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/abortions/not-a-uuid", nil)
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		remove := &fakeDeleteAbortionUseCase{}
		handler := newTestHandlerWithDelete(&fakeCreateAbortionUseCase{}, &fakeListAbortionsUseCase{}, remove)
		response := serve(handler, httptest.NewRequest(http.MethodDelete, "/api/v1/abortions/"+abortionID.String(), nil))

		if response.Code != http.StatusUnauthorized || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("maps errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"not found", ports.ErrAbortionNotFound, http.StatusNotFound},
			{"not deletable", ports.ErrAbortionNotDeletable, http.StatusConflict},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				remove := &fakeDeleteAbortionUseCase{err: test.err}
				handler := newTestHandlerWithDelete(&fakeCreateAbortionUseCase{}, &fakeListAbortionsUseCase{}, remove)
				request := httptest.NewRequest(http.MethodDelete, "/api/v1/abortions/"+uuid.New().String(), nil)
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}
