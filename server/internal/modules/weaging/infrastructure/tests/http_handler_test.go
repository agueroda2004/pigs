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

	farrowingdomain "server/internal/modules/farrowing/domain"
	weagingapplication "server/internal/modules/weaging/application"
	weagingdomain "server/internal/modules/weaging/domain"
	"server/internal/modules/weaging/ports"
)

func TestWeagingHandlerCreate(t *testing.T) {
	weaging := handlerWeaging()
	actorID := uuid.New()
	sowID := uuid.New()

	body := fmt.Sprintf(
		`{"sow_id":"%s","weaging_date":"2026-04-25","quantity":8,`+
			`"total_weight":120.5,"destination":"Nave 2","note":"destete"}`,
		sowID,
	)

	t.Run("creates a weaging with the actor from context", func(t *testing.T) {
		create := &fakeCreateWeagingUseCase{weaging: weaging}
		handler := newTestHandler(create, &fakeListWeagingsUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/weagings", strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusCreated || create.command.CreatedBy != actorID {
			t.Fatalf("status=%d command=%#v", response.Code, create.command)
		}
		if create.command.SowID != sowID || create.command.Quantity != 8 {
			t.Fatalf("unexpected command: %#v", create.command)
		}
		if create.command.TotalWeight == nil || *create.command.TotalWeight != 120.5 {
			t.Fatalf("unexpected weight: %#v", create.command.TotalWeight)
		}
		if create.command.Destination == nil || *create.command.Destination != "Nave 2" {
			t.Fatalf("unexpected destination: %#v", create.command.Destination)
		}
		if !create.command.WeagingDate.Equal(time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected date: %v", create.command.WeagingDate)
		}
		if !strings.Contains(response.Body.String(), "2026-04-25") {
			t.Fatalf("unexpected body: %s", response.Body.String())
		}
	})

	t.Run("blocks creation for non-admins", func(t *testing.T) {
		create := &fakeCreateWeagingUseCase{weaging: weaging}
		handler := newAdminGuardHandler(create)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/weagings", strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusForbidden || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		create := &fakeCreateWeagingUseCase{weaging: weaging}
		handler := newTestHandler(create, &fakeListWeagingsUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/weagings", strings.NewReader(body)))

		if response.Code != http.StatusUnauthorized || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("rejects an unknown field", func(t *testing.T) {
		create := &fakeCreateWeagingUseCase{weaging: weaging}
		handler := newTestHandler(create, &fakeListWeagingsUseCase{})
		invalid := fmt.Sprintf(
			`{"sow_id":"%s","weaging_date":"2026-04-25","quantity":8,"extra":true}`,
			sowID,
		)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/weagings", strings.NewReader(invalid))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid fields", func(t *testing.T) {
		create := &fakeCreateWeagingUseCase{weaging: weaging}
		handler := newTestHandler(create, &fakeListWeagingsUseCase{})

		for _, invalid := range []string{
			`{"sow_id":"not-a-uuid","weaging_date":"2026-04-25","quantity":8}`,
			fmt.Sprintf(`{"sow_id":"%s","weaging_date":"not-a-date","quantity":8}`, sowID),
		} {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/weagings", strings.NewReader(invalid))
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
			{"weaging not found", ports.ErrWeagingNotFound, http.StatusNotFound},
			{"sow not found", ports.ErrSowNotFound, http.StatusNotFound},
			{"farrowing not found", ports.ErrFarrowingNotFound, http.StatusNotFound},
			{"sow not lactating", weagingapplication.ErrSowNotLactating, http.StatusConflict},
			{"already exists", ports.ErrWeagingAlreadyExists, http.StatusConflict},
			{"quantity mismatch", farrowingdomain.ErrWeagingQuantityMismatch, http.StatusConflict},
			{"date before farrowing", weagingdomain.ErrWeagingDateBeforeFarrowing, http.StatusBadRequest},
			{"date before events", weagingdomain.ErrWeagingDateBeforeEvents, http.StatusBadRequest},
			{"date in future", weagingdomain.ErrWeagingDateInFuture, http.StatusBadRequest},
			{"invalid quantity", weagingdomain.ErrInvalidQuantity, http.StatusBadRequest},
			{"invalid destination", weagingdomain.ErrInvalidDestination, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				create := &fakeCreateWeagingUseCase{err: test.err}
				handler := newTestHandler(create, &fakeListWeagingsUseCase{})
				request := httptest.NewRequest(http.MethodPost, "/api/v1/weagings", strings.NewReader(body))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestWeagingHandlerList(t *testing.T) {
	t.Run("returns every weaging", func(t *testing.T) {
		list := &fakeListWeagingsUseCase{weagings: []*weagingdomain.Weaging{handlerWeaging()}}
		handler := newTestHandler(&fakeCreateWeagingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/weagings", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		body := response.Body.String()
		for _, want := range []string{"2026-04-25", "farrowing_id", "sow_id", "quantity", "total_weight", "destination"} {
			if !strings.Contains(body, want) {
				t.Fatalf("body missing %q: %s", want, body)
			}
		}
	})

	t.Run("returns an empty array when there are no weagings", func(t *testing.T) {
		list := &fakeListWeagingsUseCase{weagings: []*weagingdomain.Weaging{}}
		handler := newTestHandler(&fakeCreateWeagingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/weagings", nil))

		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("parses the sow and date range filters", func(t *testing.T) {
		sowID := uuid.New()
		list := &fakeListWeagingsUseCase{}
		handler := newTestHandler(&fakeCreateWeagingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(
			http.MethodGet,
			"/api/v1/weagings?sow_id="+sowID.String()+"&from=2026-04-01&to=2026-04-30",
			nil,
		))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		if list.filter.SowID == nil || *list.filter.SowID != sowID {
			t.Fatalf("unexpected sow id: %#v", list.filter.SowID)
		}
		if list.filter.FromDate == nil || !list.filter.FromDate.Equal(time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected from date: %#v", list.filter.FromDate)
		}
		if list.filter.ToDate == nil || !list.filter.ToDate.Equal(time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected to date: %#v", list.filter.ToDate)
		}
	})

	t.Run("returns an empty filter when no params are given", func(t *testing.T) {
		list := &fakeListWeagingsUseCase{}
		handler := newTestHandler(&fakeCreateWeagingUseCase{}, list)
		serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/weagings", nil))

		if list.filter.SowID != nil || list.filter.FromDate != nil || list.filter.ToDate != nil {
			t.Fatalf("unexpected filter: %#v", list.filter)
		}
	})

	t.Run("rejects invalid filter values", func(t *testing.T) {
		list := &fakeListWeagingsUseCase{}
		handler := newTestHandler(&fakeCreateWeagingUseCase{}, list)

		for _, query := range []string{
			"sow_id=not-a-uuid",
			"from=not-a-date",
			"to=not-a-date",
			"from=2026-04-30&to=2026-04-01",
		} {
			response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/weagings?"+query, nil))
			if response.Code != http.StatusBadRequest || list.called {
				t.Fatalf("status=%d called=%v query=%s", response.Code, list.called, query)
			}
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		list := &fakeListWeagingsUseCase{err: errors.New("unexpected")}
		handler := newTestHandler(&fakeCreateWeagingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/weagings", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}
