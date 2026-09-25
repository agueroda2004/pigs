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

	serviceapplication "server/internal/modules/service/application"
	servicedomain "server/internal/modules/service/domain"
	"server/internal/modules/service/ports"
)

func TestServiceHandlerCreate(t *testing.T) {
	service := handlerService()
	actorID := uuid.New()
	sowID := uuid.New()
	boarID := uuid.New()
	operatorID := uuid.New()

	body := fmt.Sprintf(
		`{"sow_id":"%s","location":"Nave 1","note":"ok","mounts":[{"boar_id":"%s","operator_id":"%s","mount_date":"2026-01-10","type":"Natural","note":"monta"}]}`,
		sowID, boarID, operatorID,
	)

	t.Run("creates a service with the actor from context", func(t *testing.T) {
		create := &fakeCreateServiceUseCase{service: service}
		handler := newTestHandler(create, &fakeListServicesUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/services", strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusCreated || create.command.CreatedBy != actorID {
			t.Fatalf("status=%d command=%#v", response.Code, create.command)
		}
		if create.command.SowID != sowID || len(create.command.Mounts) != 1 {
			t.Fatalf("unexpected command: %#v", create.command)
		}
		mount := create.command.Mounts[0]
		if mount.BoarID != boarID || mount.OperatorID != operatorID {
			t.Fatalf("unexpected mount: %#v", mount)
		}
		if mount.Type != servicedomain.MountTypeNatural || mount.Note == nil || *mount.Note != "monta" {
			t.Fatalf("unexpected mount: %#v", mount)
		}
		if !mount.MountDate.Equal(time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected mount date: %v", mount.MountDate)
		}
		if response.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type: %q", response.Header().Get("Content-Type"))
		}
		if !strings.Contains(response.Body.String(), "Confirmado") {
			t.Fatalf("unexpected body: %s", response.Body.String())
		}
	})

	t.Run("leaves the type empty so the domain defaults it", func(t *testing.T) {
		create := &fakeCreateServiceUseCase{service: service}
		handler := newTestHandler(create, &fakeListServicesUseCase{})
		withoutType := fmt.Sprintf(
			`{"sow_id":"%s","mounts":[{"boar_id":"%s","operator_id":"%s","mount_date":"2026-01-10"}]}`,
			sowID, boarID, operatorID,
		)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/services", strings.NewReader(withoutType))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusCreated || create.command.Mounts[0].Type != "" {
			t.Fatalf("status=%d type=%q", response.Code, create.command.Mounts[0].Type)
		}
	})

	t.Run("rejects an unknown field", func(t *testing.T) {
		create := &fakeCreateServiceUseCase{service: service}
		handler := newTestHandler(create, &fakeListServicesUseCase{})
		invalid := fmt.Sprintf(
			`{"sow_id":"%s","state":"Confirmado","mounts":[]}`,
			sowID,
		)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/services", strings.NewReader(invalid))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		create := &fakeCreateServiceUseCase{service: service}
		handler := newTestHandler(create, &fakeListServicesUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/services", strings.NewReader(body)))

		if response.Code != http.StatusUnauthorized || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid JSON", func(t *testing.T) {
		create := &fakeCreateServiceUseCase{service: service}
		handler := newTestHandler(create, &fakeListServicesUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/services", strings.NewReader(`{"sow_id":`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid identifiers and dates", func(t *testing.T) {
		create := &fakeCreateServiceUseCase{service: service}
		handler := newTestHandler(create, &fakeListServicesUseCase{})

		for _, invalid := range []string{
			`{"sow_id":"not-a-uuid","mounts":[]}`,
			fmt.Sprintf(`{"sow_id":"%s","mounts":[{"boar_id":"not-a-uuid","operator_id":"%s","mount_date":"2026-01-10"}]}`, sowID, operatorID),
			fmt.Sprintf(`{"sow_id":"%s","mounts":[{"boar_id":"%s","operator_id":"not-a-uuid","mount_date":"2026-01-10"}]}`, sowID, boarID),
			fmt.Sprintf(`{"sow_id":"%s","mounts":[{"boar_id":"%s","operator_id":"%s","mount_date":"not-a-date"}]}`, sowID, boarID, operatorID),
		} {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/services", strings.NewReader(invalid))
			request = authenticatedRequest(request, actorID)
			response := serve(handler, request)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", response.Code, invalid)
			}
		}
	})

	t.Run("maps application and domain errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"sow not found", ports.ErrSowNotFound, http.StatusNotFound},
			{"boar not found", ports.ErrBoarNotFound, http.StatusNotFound},
			{"operator not found", ports.ErrOperatorNotFound, http.StatusNotFound},
			{"sow not eligible", serviceapplication.ErrSowNotEligible, http.StatusConflict},
			{"boar not eligible", serviceapplication.ErrBoarNotEligible, http.StatusConflict},
			{"operator not available", serviceapplication.ErrOperatorNotAvailable, http.StatusConflict},
			{"invalid mounts", servicedomain.ErrInvalidMounts, http.StatusBadRequest},
			{"mount in future", servicedomain.ErrMountDateInFuture, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				create := &fakeCreateServiceUseCase{err: test.err}
				handler := newTestHandler(create, &fakeListServicesUseCase{})
				request := httptest.NewRequest(http.MethodPost, "/api/v1/services", strings.NewReader(body))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestServiceHandlerList(t *testing.T) {
	t.Run("returns every service with its mounts", func(t *testing.T) {
		list := &fakeListServicesUseCase{services: []*servicedomain.Service{handlerService()}}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/services", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		body := response.Body.String()
		for _, want := range []string{"Confirmado", "2026-01-10", "Artificial", "mounts"} {
			if !strings.Contains(body, want) {
				t.Fatalf("body missing %q: %s", want, body)
			}
		}
	})

	t.Run("returns an empty array when there are no services", func(t *testing.T) {
		list := &fakeListServicesUseCase{services: []*servicedomain.Service{}}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/services", nil))

		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("parses the query filters", func(t *testing.T) {
		sowID := uuid.New()
		list := &fakeListServicesUseCase{}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, list)
		target := "/api/v1/services?sow_id=" + sowID.String() + "&state=Fallido"
		response := serve(handler, httptest.NewRequest(http.MethodGet, target, nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		if list.filter.SowID == nil || *list.filter.SowID != sowID {
			t.Fatalf("unexpected sow id: %#v", list.filter.SowID)
		}
		if list.filter.State == nil || *list.filter.State != servicedomain.StateFailed {
			t.Fatalf("unexpected state: %#v", list.filter.State)
		}
	})

	t.Run("returns an empty filter when no params are given", func(t *testing.T) {
		list := &fakeListServicesUseCase{}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, list)
		serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/services", nil))

		if list.filter.SowID != nil || list.filter.State != nil {
			t.Fatalf("unexpected filter: %#v", list.filter)
		}
	})

	t.Run("rejects invalid filter values", func(t *testing.T) {
		list := &fakeListServicesUseCase{}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, list)

		for _, target := range []string{
			"/api/v1/services?sow_id=not-a-uuid",
			"/api/v1/services?state=Desconocido",
		} {
			response := serve(handler, httptest.NewRequest(http.MethodGet, target, nil))
			if response.Code != http.StatusBadRequest || list.called {
				t.Fatalf("target=%s status=%d called=%v", target, response.Code, list.called)
			}
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		list := &fakeListServicesUseCase{err: errors.New("unexpected")}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/services", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}
