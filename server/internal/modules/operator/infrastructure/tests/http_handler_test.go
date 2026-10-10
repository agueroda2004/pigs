package tests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	operatordomain "server/internal/modules/operator/domain"
	"server/internal/modules/operator/ports"
)

func TestOperatorHandlerCreate(t *testing.T) {
	operator := handlerOperator()
	actorID := uuid.New()

	t.Run("creates an operator with actor from context", func(t *testing.T) {
		create := &fakeCreateOperatorUseCase{operator: operator}
		handler := newTestHandler(create, &fakeListOperatorsUseCase{}, &fakeUpdateOperatorUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/operators", strings.NewReader(`{"name":"Juan Pérez"}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusCreated || create.command.CreatedBy != actorID || create.command.Name != "Juan Pérez" {
			t.Fatalf("status=%d command=%#v", response.Code, create.command)
		}
		if response.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type: %q", response.Header().Get("Content-Type"))
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		create := &fakeCreateOperatorUseCase{operator: operator}
		handler := newTestHandler(create, &fakeListOperatorsUseCase{}, &fakeUpdateOperatorUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/operators", strings.NewReader(`{"name":"Juan Pérez"}`)))

		if response.Code != http.StatusUnauthorized || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid JSON", func(t *testing.T) {
		create := &fakeCreateOperatorUseCase{operator: operator}
		handler := newTestHandler(create, &fakeListOperatorsUseCase{}, &fakeUpdateOperatorUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/operators", strings.NewReader(`{"name":`)))

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
			{"duplicate", ports.ErrOperatorNameAlreadyUsed, http.StatusConflict},
			{"validation", operatordomain.ErrInvalidName, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				create := &fakeCreateOperatorUseCase{err: test.err}
				handler := newTestHandler(create, &fakeListOperatorsUseCase{}, &fakeUpdateOperatorUseCase{})
				request := httptest.NewRequest(http.MethodPost, "/api/v1/operators", strings.NewReader(`{"name":"Juan Pérez"}`))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestOperatorHandlerList(t *testing.T) {
	t.Run("returns every operator", func(t *testing.T) {
		list := &fakeListOperatorsUseCase{operators: []*operatordomain.Operator{handlerOperator(), handlerOperator()}}
		handler := newTestHandler(&fakeCreateOperatorUseCase{}, list, &fakeUpdateOperatorUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/operators", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		body := strings.TrimSpace(response.Body.String())
		if !strings.HasPrefix(body, "[") || !strings.Contains(body, "Juan Pérez") {
			t.Fatalf("unexpected body: %s", body)
		}
	})

	t.Run("returns an empty array when there are no operators", func(t *testing.T) {
		list := &fakeListOperatorsUseCase{operators: []*operatordomain.Operator{}}
		handler := newTestHandler(&fakeCreateOperatorUseCase{}, list, &fakeUpdateOperatorUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/operators", nil))

		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		list := &fakeListOperatorsUseCase{err: errors.New("unexpected")}
		handler := newTestHandler(&fakeCreateOperatorUseCase{}, list, &fakeUpdateOperatorUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/operators", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}

func TestOperatorHandlerListDropdown(t *testing.T) {
	options := []operatordomain.OperatorDropdown{
		{ID: uuid.New(), Name: "Ana", Active: true},
		{ID: uuid.New(), Name: "Beto", Active: false},
	}

	t.Run("returns the dropdown items and forwards the active filter", func(t *testing.T) {
		dropdown := &fakeListOperatorDropdownUseCase{options: options}
		handler := newTestHandlerWithDropdown(&fakeCreateOperatorUseCase{}, &fakeListOperatorsUseCase{}, dropdown, &fakeUpdateOperatorUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/operators/dropdown?active=true", nil))

		if response.Code != http.StatusOK || !dropdown.called {
			t.Fatalf("status=%d called=%v", response.Code, dropdown.called)
		}
		if dropdown.active == nil || !*dropdown.active {
			t.Fatalf("unexpected active filter: %#v", dropdown.active)
		}
		body := response.Body.String()
		for _, want := range []string{"Ana", "Beto", `"active"`} {
			if !strings.Contains(body, want) {
				t.Fatalf("body missing %q: %s", want, body)
			}
		}
	})

	t.Run("returns every operator when no active filter is given", func(t *testing.T) {
		dropdown := &fakeListOperatorDropdownUseCase{options: options}
		handler := newTestHandlerWithDropdown(&fakeCreateOperatorUseCase{}, &fakeListOperatorsUseCase{}, dropdown, &fakeUpdateOperatorUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/operators/dropdown", nil))

		if response.Code != http.StatusOK || dropdown.active != nil {
			t.Fatalf("status=%d active=%#v", response.Code, dropdown.active)
		}
	})

	t.Run("rejects a malformed active filter", func(t *testing.T) {
		dropdown := &fakeListOperatorDropdownUseCase{}
		handler := newTestHandlerWithDropdown(&fakeCreateOperatorUseCase{}, &fakeListOperatorsUseCase{}, dropdown, &fakeUpdateOperatorUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/operators/dropdown?active=maybe", nil))

		if response.Code != http.StatusBadRequest || dropdown.called {
			t.Fatalf("status=%d called=%v", response.Code, dropdown.called)
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		dropdown := &fakeListOperatorDropdownUseCase{err: errors.New("unexpected")}
		handler := newTestHandlerWithDropdown(&fakeCreateOperatorUseCase{}, &fakeListOperatorsUseCase{}, dropdown, &fakeUpdateOperatorUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/operators/dropdown", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}

func TestOperatorHandlerUpdate(t *testing.T) {
	operatorID := uuid.New()
	actorID := uuid.New()

	t.Run("updates the provided fields and forwards the actor", func(t *testing.T) {
		update := &fakeUpdateOperatorUseCase{operator: handlerOperator()}
		handler := newTestHandler(&fakeCreateOperatorUseCase{}, &fakeListOperatorsUseCase{}, update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/operators/"+operatorID.String(), strings.NewReader(`{"name":"New Name","active":false}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || update.operatorID != operatorID || update.command.UpdatedBy != actorID || update.command.Name == nil || *update.command.Name != "New Name" || update.command.Active == nil || *update.command.Active {
			t.Fatalf("status=%d id=%v command=%#v", response.Code, update.operatorID, update.command)
		}
	})

	t.Run("returns bad request for invalid UUID", func(t *testing.T) {
		update := &fakeUpdateOperatorUseCase{}
		handler := newTestHandler(&fakeCreateOperatorUseCase{}, &fakeListOperatorsUseCase{}, update)
		response := serve(handler, httptest.NewRequest(http.MethodPatch, "/api/v1/operators/not-a-uuid", strings.NewReader(`{"name":"New"}`)))

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		update := &fakeUpdateOperatorUseCase{operator: handlerOperator()}
		handler := newTestHandler(&fakeCreateOperatorUseCase{}, &fakeListOperatorsUseCase{}, update)
		response := serve(handler, httptest.NewRequest(http.MethodPatch, "/api/v1/operators/"+operatorID.String(), strings.NewReader(`{"name":"New"}`)))

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
			{"not found", ports.ErrOperatorNotFound, http.StatusNotFound},
			{"duplicate", ports.ErrOperatorNameAlreadyUsed, http.StatusConflict},
			{"validation", operatordomain.ErrInvalidUpdate, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				update := &fakeUpdateOperatorUseCase{err: test.err}
				handler := newTestHandler(&fakeCreateOperatorUseCase{}, &fakeListOperatorsUseCase{}, update)
				request := httptest.NewRequest(http.MethodPatch, "/api/v1/operators/"+uuid.New().String(), strings.NewReader(`{"name":"New"}`))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}
