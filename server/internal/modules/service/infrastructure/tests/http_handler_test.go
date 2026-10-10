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
		handler := newTestHandler(create, &fakeListServicesUseCase{}, &fakeDeleteServiceUseCase{}, &fakeUpdateServiceUseCase{})
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
		if strings.TrimSpace(response.Body.String()) != "" {
			t.Fatalf("expected an empty body, got: %s", response.Body.String())
		}
	})

	t.Run("leaves the type empty so the domain defaults it", func(t *testing.T) {
		create := &fakeCreateServiceUseCase{service: service}
		handler := newTestHandler(create, &fakeListServicesUseCase{}, &fakeDeleteServiceUseCase{}, &fakeUpdateServiceUseCase{})
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
		handler := newTestHandler(create, &fakeListServicesUseCase{}, &fakeDeleteServiceUseCase{}, &fakeUpdateServiceUseCase{})
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
		handler := newTestHandler(create, &fakeListServicesUseCase{}, &fakeDeleteServiceUseCase{}, &fakeUpdateServiceUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/services", strings.NewReader(body)))

		if response.Code != http.StatusUnauthorized || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid JSON", func(t *testing.T) {
		create := &fakeCreateServiceUseCase{service: service}
		handler := newTestHandler(create, &fakeListServicesUseCase{}, &fakeDeleteServiceUseCase{}, &fakeUpdateServiceUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/services", strings.NewReader(`{"sow_id":`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid identifiers and dates", func(t *testing.T) {
		create := &fakeCreateServiceUseCase{service: service}
		handler := newTestHandler(create, &fakeListServicesUseCase{}, &fakeDeleteServiceUseCase{}, &fakeUpdateServiceUseCase{})

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
				handler := newTestHandler(create, &fakeListServicesUseCase{}, &fakeDeleteServiceUseCase{}, &fakeUpdateServiceUseCase{})
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
	t.Run("returns the paginated services without audit fields", func(t *testing.T) {
		list := &fakeListServicesUseCase{services: []*servicedomain.Service{handlerService()}}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, list, &fakeDeleteServiceUseCase{}, &fakeUpdateServiceUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/services", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		body := response.Body.String()
		for _, want := range []string{
			"items", "Confirmado", "2026-01-10", "Artificial", "mounts",
			"sow_code", "C-001", "boar_code", "V-001", "operator_name", "Ana",
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("body missing %q: %s", want, body)
			}
		}
		if strings.Contains(body, "created_at") || strings.Contains(body, "updated_at") ||
			strings.Contains(body, "created_by") || strings.Contains(body, "updated_by") {
			t.Fatalf("list must omit audit fields: %s", body)
		}
	})

	t.Run("returns an empty items array when there are no services", func(t *testing.T) {
		list := &fakeListServicesUseCase{services: []*servicedomain.Service{}}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, list, &fakeDeleteServiceUseCase{}, &fakeUpdateServiceUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/services", nil))

		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"items":[]`) {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("parses the page query parameter", func(t *testing.T) {
		list := &fakeListServicesUseCase{}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, list, &fakeDeleteServiceUseCase{}, &fakeUpdateServiceUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/services?page=3", nil))

		if response.Code != http.StatusOK || list.page != 3 {
			t.Fatalf("status=%d page=%d", response.Code, list.page)
		}
	})

	t.Run("rejects an invalid page", func(t *testing.T) {
		list := &fakeListServicesUseCase{}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, list, &fakeDeleteServiceUseCase{}, &fakeUpdateServiceUseCase{})

		for _, target := range []string{"/api/v1/services?page=0", "/api/v1/services?page=abc"} {
			response := serve(handler, httptest.NewRequest(http.MethodGet, target, nil))
			if response.Code != http.StatusBadRequest || list.called {
				t.Fatalf("target=%s status=%d called=%v", target, response.Code, list.called)
			}
		}
	})

	t.Run("parses the query filters", func(t *testing.T) {
		list := &fakeListServicesUseCase{}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, list, &fakeDeleteServiceUseCase{}, &fakeUpdateServiceUseCase{})
		target := "/api/v1/services?sow_code=C-001&state=Fallido"
		response := serve(handler, httptest.NewRequest(http.MethodGet, target, nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		if list.filter.SowCode == nil || *list.filter.SowCode != "C-001" {
			t.Fatalf("unexpected sow code: %#v", list.filter.SowCode)
		}
		if list.filter.State == nil || *list.filter.State != servicedomain.StateFailed {
			t.Fatalf("unexpected state: %#v", list.filter.State)
		}
	})

	t.Run("trims the sow code filter", func(t *testing.T) {
		list := &fakeListServicesUseCase{}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, list, &fakeDeleteServiceUseCase{}, &fakeUpdateServiceUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/services?sow_code=%20C-001%20", nil))

		if response.Code != http.StatusOK || list.filter.SowCode == nil || *list.filter.SowCode != "C-001" {
			t.Fatalf("status=%d filter=%#v", response.Code, list.filter.SowCode)
		}
	})

	t.Run("accepts a sow code at the length limit", func(t *testing.T) {
		list := &fakeListServicesUseCase{}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, list, &fakeDeleteServiceUseCase{}, &fakeUpdateServiceUseCase{})
		code := strings.Repeat("A", 50)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/services?sow_code="+code, nil))

		if response.Code != http.StatusOK || list.filter.SowCode == nil || *list.filter.SowCode != code {
			t.Fatalf("status=%d filter=%#v", response.Code, list.filter.SowCode)
		}
	})

	t.Run("returns an empty filter when no params are given", func(t *testing.T) {
		list := &fakeListServicesUseCase{}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, list, &fakeDeleteServiceUseCase{}, &fakeUpdateServiceUseCase{})
		serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/services", nil))

		if list.filter.SowCode != nil || list.filter.State != nil {
			t.Fatalf("unexpected filter: %#v", list.filter)
		}
	})

	t.Run("rejects invalid filter values", func(t *testing.T) {
		list := &fakeListServicesUseCase{}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, list, &fakeDeleteServiceUseCase{}, &fakeUpdateServiceUseCase{})

		for _, target := range []string{
			"/api/v1/services?state=Desconocido",
			"/api/v1/services?sow_code=" + strings.Repeat("A", 51),
		} {
			response := serve(handler, httptest.NewRequest(http.MethodGet, target, nil))
			if response.Code != http.StatusBadRequest || list.called {
				t.Fatalf("target=%s status=%d called=%v", target, response.Code, list.called)
			}
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		list := &fakeListServicesUseCase{err: errors.New("unexpected")}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, list, &fakeDeleteServiceUseCase{}, &fakeUpdateServiceUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/services", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}

func TestServiceHandlerDelete(t *testing.T) {
	serviceID := uuid.New()
	actorID := uuid.New()

	t.Run("deletes the service with the provided id", func(t *testing.T) {
		remove := &fakeDeleteServiceUseCase{}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, &fakeListServicesUseCase{}, remove, &fakeUpdateServiceUseCase{})
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/services/"+serviceID.String(), nil)
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusNoContent || !remove.called ||
			remove.command.ID != serviceID || remove.command.DeletedBy != actorID {
			t.Fatalf("status=%d command=%#v", response.Code, remove.command)
		}
		if strings.TrimSpace(response.Body.String()) != "" {
			t.Fatalf("expected an empty body, got: %s", response.Body.String())
		}
	})

	t.Run("returns bad request for invalid UUID", func(t *testing.T) {
		remove := &fakeDeleteServiceUseCase{}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, &fakeListServicesUseCase{}, remove, &fakeUpdateServiceUseCase{})
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/services/not-a-uuid", nil)
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		remove := &fakeDeleteServiceUseCase{}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, &fakeListServicesUseCase{}, remove, &fakeUpdateServiceUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodDelete, "/api/v1/services/"+serviceID.String(), nil))

		if response.Code != http.StatusUnauthorized || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("blocks deletion for non-admins", func(t *testing.T) {
		remove := &fakeDeleteServiceUseCase{}
		handler := newAdminGuardHandler(remove, &fakeUpdateServiceUseCase{})
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/services/"+serviceID.String(), nil)
		request = authenticatedRequest(request, actorID)
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
			{"not found", ports.ErrServiceNotFound, http.StatusNotFound},
			{"not deletable", servicedomain.ErrServiceNotDeletable, http.StatusConflict},
			{"in use", ports.ErrServiceInUse, http.StatusConflict},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				remove := &fakeDeleteServiceUseCase{err: test.err}
				handler := newTestHandler(&fakeCreateServiceUseCase{}, &fakeListServicesUseCase{}, remove, &fakeUpdateServiceUseCase{})
				request := httptest.NewRequest(http.MethodDelete, "/api/v1/services/"+uuid.New().String(), nil)
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestServiceHandlerUpdate(t *testing.T) {
	serviceID := uuid.New()
	actorID := uuid.New()
	boarID := uuid.New()
	operatorID := uuid.New()
	mountID := uuid.New()

	body := fmt.Sprintf(
		`{"location":"Nave 2","note":"ok","mounts":{"create":[{"boar_id":"%s","operator_id":"%s","mount_date":"2026-01-10","type":"Natural"}],"update":[{"id":"%s","boar_id":"%s","operator_id":"%s","mount_date":"2026-01-11","type":"Artificial"}],"delete":["%s"]}}`,
		boarID, operatorID, mountID, boarID, operatorID, uuid.New(),
	)

	t.Run("applies the update and returns no body", func(t *testing.T) {
		update := &fakeUpdateServiceUseCase{}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, &fakeListServicesUseCase{}, &fakeDeleteServiceUseCase{}, update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/services/"+serviceID.String(), strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || !update.called || update.id != serviceID {
			t.Fatalf("status=%d id=%v command=%#v", response.Code, update.id, update.command)
		}
		if update.command.UpdatedBy != actorID || len(update.command.CreateMounts) != 1 {
			t.Fatalf("unexpected command: %#v", update.command)
		}
		if len(update.command.UpdateMounts) != 1 || update.command.UpdateMounts[0].ID != mountID {
			t.Fatalf("unexpected update mounts: %#v", update.command.UpdateMounts)
		}
		if len(update.command.DeleteMountIDs) != 1 {
			t.Fatalf("unexpected delete mounts: %#v", update.command.DeleteMountIDs)
		}
		if strings.TrimSpace(response.Body.String()) != "" {
			t.Fatalf("expected an empty body, got: %s", response.Body.String())
		}
	})

	t.Run("returns bad request for invalid identifiers", func(t *testing.T) {
		update := &fakeUpdateServiceUseCase{}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, &fakeListServicesUseCase{}, &fakeDeleteServiceUseCase{}, update)

		request := httptest.NewRequest(http.MethodPatch, "/api/v1/services/not-a-uuid", strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)
		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}

		invalid := `{"mounts":{"create":[{"boar_id":"not-a-uuid","operator_id":"x","mount_date":"2026-01-10"}]}}`
		request = httptest.NewRequest(http.MethodPatch, "/api/v1/services/"+serviceID.String(), strings.NewReader(invalid))
		request = authenticatedRequest(request, actorID)
		response = serve(handler, request)
		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		update := &fakeUpdateServiceUseCase{}
		handler := newTestHandler(&fakeCreateServiceUseCase{}, &fakeListServicesUseCase{}, &fakeDeleteServiceUseCase{}, update)
		response := serve(handler, httptest.NewRequest(http.MethodPatch, "/api/v1/services/"+serviceID.String(), strings.NewReader(body)))

		if response.Code != http.StatusUnauthorized || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("blocks update for non-admins", func(t *testing.T) {
		update := &fakeUpdateServiceUseCase{}
		handler := newAdminGuardHandler(&fakeDeleteServiceUseCase{}, update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/services/"+serviceID.String(), strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusForbidden || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("maps application errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"not found", ports.ErrServiceNotFound, http.StatusNotFound},
			{"not editable", servicedomain.ErrServiceNotEditable, http.StatusConflict},
			{"before previous service", servicedomain.ErrMountBeforePreviousService, http.StatusBadRequest},
			{"mount not found", servicedomain.ErrMountNotFound, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				update := &fakeUpdateServiceUseCase{err: test.err}
				handler := newTestHandler(&fakeCreateServiceUseCase{}, &fakeListServicesUseCase{}, &fakeDeleteServiceUseCase{}, update)
				request := httptest.NewRequest(http.MethodPatch, "/api/v1/services/"+uuid.New().String(), strings.NewReader(body))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}
