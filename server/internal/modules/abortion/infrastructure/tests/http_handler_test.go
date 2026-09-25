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
		if response.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type: %q", response.Header().Get("Content-Type"))
		}
		if !strings.Contains(response.Body.String(), "Infeccioso") {
			t.Fatalf("unexpected body: %s", response.Body.String())
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
		for _, want := range []string{"Infeccioso", "2026-01-15", "service_id"} {
			if !strings.Contains(body, want) {
				t.Fatalf("body missing %q: %s", want, body)
			}
		}
	})

	t.Run("returns an empty array when there are no abortions", func(t *testing.T) {
		list := &fakeListAbortionsUseCase{abortions: []*abortiondomain.Abortion{}}
		handler := newTestHandler(&fakeCreateAbortionUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/abortions", nil))

		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("parses the sow_id filter", func(t *testing.T) {
		sowID := uuid.New()
		list := &fakeListAbortionsUseCase{}
		handler := newTestHandler(&fakeCreateAbortionUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/abortions?sow_id="+sowID.String(), nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		if list.filter.SowID == nil || *list.filter.SowID != sowID {
			t.Fatalf("unexpected sow id: %#v", list.filter.SowID)
		}
	})

	t.Run("returns an empty filter when no params are given", func(t *testing.T) {
		list := &fakeListAbortionsUseCase{}
		handler := newTestHandler(&fakeCreateAbortionUseCase{}, list)
		serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/abortions", nil))

		if list.filter.SowID != nil {
			t.Fatalf("unexpected filter: %#v", list.filter)
		}
	})

	t.Run("rejects invalid filter values", func(t *testing.T) {
		list := &fakeListAbortionsUseCase{}
		handler := newTestHandler(&fakeCreateAbortionUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/abortions?sow_id=not-a-uuid", nil))

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
