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
		if response.Body.Len() != 0 {
			t.Fatalf("body=%q, want empty", response.Body.String())
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
			{"quantity above max", weagingdomain.ErrQuantityExceedsMax, http.StatusBadRequest},
			{"invalid total weight", weagingdomain.ErrInvalidTotalWeight, http.StatusBadRequest},
			{"total weight above max", weagingdomain.ErrTotalWeightExceedsMax, http.StatusBadRequest},
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
	t.Run("returns one page with metadata and sow code", func(t *testing.T) {
		list := &fakeListWeagingsUseCase{page: weagingapplication.WeagingPage{
			Items:      []*weagingdomain.Weaging{handlerWeaging()},
			Total:      12,
			Page:       2,
			PageSize:   10,
			TotalPages: 2,
		}}
		handler := newTestHandler(&fakeCreateWeagingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/weagings?page=2", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		if list.pageArg != 2 {
			t.Fatalf("page=%d, want 2", list.pageArg)
		}
		body := response.Body.String()
		for _, want := range []string{
			`"items"`,
			`"total":12`,
			`"page":2`,
			`"page_size":10`,
			`"total_pages":2`,
			`"sow_code":"C-001"`,
			`"farrowing_id"`,
			`"sow_id"`,
			`"quantity"`,
			`"total_weight"`,
			`"destination"`,
			`2026-04-25`,
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("body missing %q: %s", want, body)
			}
		}
	})

	t.Run("omits the audit fields", func(t *testing.T) {
		list := &fakeListWeagingsUseCase{page: weagingapplication.WeagingPage{
			Items: []*weagingdomain.Weaging{handlerWeaging()},
		}}
		handler := newTestHandler(&fakeCreateWeagingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/weagings", nil))

		body := response.Body.String()
		for _, unwanted := range []string{"created_at", "updated_at", "created_by", "updated_by"} {
			if strings.Contains(body, unwanted) {
				t.Fatalf("body should omit %q: %s", unwanted, body)
			}
		}
	})

	t.Run("defaults to page one and an empty items array", func(t *testing.T) {
		list := &fakeListWeagingsUseCase{}
		handler := newTestHandler(&fakeCreateWeagingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/weagings", nil))

		if response.Code != http.StatusOK || list.pageArg != 1 {
			t.Fatalf("status=%d page=%d", response.Code, list.pageArg)
		}
		if !strings.Contains(response.Body.String(), `"items":[]`) {
			t.Fatalf("unexpected body: %s", response.Body.String())
		}
	})

	t.Run("parses the sow code and date range filters", func(t *testing.T) {
		list := &fakeListWeagingsUseCase{}
		handler := newTestHandler(&fakeCreateWeagingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(
			http.MethodGet,
			"/api/v1/weagings?sow_code=%20C-001%20&from=2026-04-01&to=2026-04-30",
			nil,
		))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		if list.filter.SowCode == nil || *list.filter.SowCode != "C-001" {
			t.Fatalf("unexpected sow code: %#v", list.filter.SowCode)
		}
		if list.filter.FromDate == nil || !list.filter.FromDate.Equal(time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected from date: %#v", list.filter.FromDate)
		}
		if list.filter.ToDate == nil || !list.filter.ToDate.Equal(time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected to date: %#v", list.filter.ToDate)
		}
	})

	t.Run("accepts a 50-character sow code", func(t *testing.T) {
		sowCode := strings.Repeat("a", 50)
		list := &fakeListWeagingsUseCase{}
		handler := newTestHandler(&fakeCreateWeagingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(
			http.MethodGet,
			"/api/v1/weagings?sow_code="+sowCode,
			nil,
		))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		if list.filter.SowCode == nil || *list.filter.SowCode != sowCode {
			t.Fatalf("unexpected sow code: %#v", list.filter.SowCode)
		}
	})

	t.Run("rejects a sow code longer than 50 characters", func(t *testing.T) {
		list := &fakeListWeagingsUseCase{}
		handler := newTestHandler(&fakeCreateWeagingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(
			http.MethodGet,
			"/api/v1/weagings?sow_code="+strings.Repeat("a", 51),
			nil,
		))

		if response.Code != http.StatusBadRequest || list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
	})

	t.Run("returns an empty filter when no params are given", func(t *testing.T) {
		list := &fakeListWeagingsUseCase{}
		handler := newTestHandler(&fakeCreateWeagingUseCase{}, list)
		serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/weagings", nil))

		if list.filter.SowCode != nil || list.filter.FromDate != nil || list.filter.ToDate != nil {
			t.Fatalf("unexpected filter: %#v", list.filter)
		}
	})

	t.Run("rejects invalid filter and page values", func(t *testing.T) {
		list := &fakeListWeagingsUseCase{}
		handler := newTestHandler(&fakeCreateWeagingUseCase{}, list)

		for _, query := range []string{
			"from=not-a-date",
			"to=not-a-date",
			"from=2026-04-30&to=2026-04-01",
			"page=not-a-number",
			"page=0",
			"page=-1",
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

func TestWeagingHandlerUpdate(t *testing.T) {
	actorID := uuid.New()

	t.Run("updates the weaging and returns no body", func(t *testing.T) {
		weaging := handlerWeaging()
		update := &fakeUpdateWeagingUseCase{}
		handler := newTestHandlerWithUpdate(&fakeCreateWeagingUseCase{}, update, &fakeListWeagingsUseCase{})
		body := `{"total_weight":130.5,"destination":"Nave 3","note":"nueva nota"}`
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/weagings/"+weaging.ID.String(), strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || !update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
		if response.Body.Len() != 0 {
			t.Fatalf("body=%q, want empty", response.Body.String())
		}
		if update.id != weaging.ID || update.command.UpdatedBy != actorID {
			t.Fatalf("unexpected command: %#v", update)
		}
		if update.command.TotalWeight == nil || *update.command.TotalWeight != 130.5 {
			t.Fatalf("unexpected weight: %#v", update.command.TotalWeight)
		}
		if update.command.Destination == nil || *update.command.Destination != "Nave 3" {
			t.Fatalf("unexpected destination: %#v", update.command.Destination)
		}
		if update.command.Note == nil || *update.command.Note != "nueva nota" {
			t.Fatalf("unexpected note: %#v", update.command.Note)
		}
		if update.command.ClearTotalWeight {
			t.Fatalf("clear weight should be false")
		}
	})

	t.Run("passes the clear-weight flag", func(t *testing.T) {
		weaging := handlerWeaging()
		update := &fakeUpdateWeagingUseCase{}
		handler := newTestHandlerWithUpdate(&fakeCreateWeagingUseCase{}, update, &fakeListWeagingsUseCase{})
		request := httptest.NewRequest(
			http.MethodPatch,
			"/api/v1/weagings/"+weaging.ID.String(),
			strings.NewReader(`{"clear_total_weight":true}`),
		)
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || !update.command.ClearTotalWeight {
			t.Fatalf("status=%d command=%#v", response.Code, update.command)
		}
		if update.command.TotalWeight != nil {
			t.Fatalf("weight should be nil: %#v", update.command.TotalWeight)
		}
	})

	t.Run("rejects an invalid id", func(t *testing.T) {
		update := &fakeUpdateWeagingUseCase{}
		handler := newTestHandlerWithUpdate(&fakeCreateWeagingUseCase{}, update, &fakeListWeagingsUseCase{})
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/weagings/not-a-uuid", strings.NewReader(`{}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("rejects malformed JSON", func(t *testing.T) {
		weaging := handlerWeaging()
		update := &fakeUpdateWeagingUseCase{}
		handler := newTestHandlerWithUpdate(&fakeCreateWeagingUseCase{}, update, &fakeListWeagingsUseCase{})
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/weagings/"+weaging.ID.String(), strings.NewReader(`{"total_weight":`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		weaging := handlerWeaging()
		update := &fakeUpdateWeagingUseCase{}
		handler := newTestHandlerWithUpdate(&fakeCreateWeagingUseCase{}, update, &fakeListWeagingsUseCase{})
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/weagings/"+weaging.ID.String(), strings.NewReader(`{}`))
		response := serve(handler, request)

		if response.Code != http.StatusUnauthorized || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("blocks the update for non-admins", func(t *testing.T) {
		weaging := handlerWeaging()
		update := &fakeUpdateWeagingUseCase{}
		handler := newAdminGuardUpdateHandler(update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/weagings/"+weaging.ID.String(), strings.NewReader(`{}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusForbidden || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("maps a missing weaging", func(t *testing.T) {
		weaging := handlerWeaging()
		update := &fakeUpdateWeagingUseCase{err: ports.ErrWeagingNotFound}
		handler := newTestHandlerWithUpdate(&fakeCreateWeagingUseCase{}, update, &fakeListWeagingsUseCase{})
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/weagings/"+weaging.ID.String(), strings.NewReader(`{}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusNotFound {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusNotFound)
		}
	})

	t.Run("maps an empty update to a bad request", func(t *testing.T) {
		weaging := handlerWeaging()
		update := &fakeUpdateWeagingUseCase{err: weagingdomain.ErrInvalidUpdate}
		handler := newTestHandlerWithUpdate(&fakeCreateWeagingUseCase{}, update, &fakeListWeagingsUseCase{})
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/weagings/"+weaging.ID.String(), strings.NewReader(`{}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusBadRequest)
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		weaging := handlerWeaging()
		update := &fakeUpdateWeagingUseCase{err: errors.New("unexpected")}
		handler := newTestHandlerWithUpdate(&fakeCreateWeagingUseCase{}, update, &fakeListWeagingsUseCase{})
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/weagings/"+weaging.ID.String(), strings.NewReader(`{}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}

func TestWeagingHandlerDelete(t *testing.T) {
	actorID := uuid.New()

	t.Run("deletes the weaging and returns no content", func(t *testing.T) {
		weaging := handlerWeaging()
		remove := &fakeDeleteWeagingUseCase{}
		handler := newTestHandlerWithDelete(&fakeCreateWeagingUseCase{}, &fakeUpdateWeagingUseCase{}, remove, &fakeListWeagingsUseCase{})
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/weagings/"+weaging.ID.String(), nil)
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusNoContent || !remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
		if response.Body.Len() != 0 {
			t.Fatalf("body=%q, want empty", response.Body.String())
		}
		if remove.command.ID != weaging.ID || remove.command.DeletedBy != actorID {
			t.Fatalf("unexpected command: %#v", remove.command)
		}
	})

	t.Run("rejects an invalid id", func(t *testing.T) {
		remove := &fakeDeleteWeagingUseCase{}
		handler := newTestHandlerWithDelete(&fakeCreateWeagingUseCase{}, &fakeUpdateWeagingUseCase{}, remove, &fakeListWeagingsUseCase{})
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/weagings/not-a-uuid", nil)
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		weaging := handlerWeaging()
		remove := &fakeDeleteWeagingUseCase{}
		handler := newTestHandlerWithDelete(&fakeCreateWeagingUseCase{}, &fakeUpdateWeagingUseCase{}, remove, &fakeListWeagingsUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodDelete, "/api/v1/weagings/"+weaging.ID.String(), nil))

		if response.Code != http.StatusUnauthorized || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("blocks deletion for non-admins", func(t *testing.T) {
		weaging := handlerWeaging()
		remove := &fakeDeleteWeagingUseCase{}
		handler := newAdminGuardDeleteHandler(remove)
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/weagings/"+weaging.ID.String(), nil)
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusForbidden || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("maps a missing weaging", func(t *testing.T) {
		weaging := handlerWeaging()
		remove := &fakeDeleteWeagingUseCase{err: ports.ErrWeagingNotFound}
		handler := newTestHandlerWithDelete(&fakeCreateWeagingUseCase{}, &fakeUpdateWeagingUseCase{}, remove, &fakeListWeagingsUseCase{})
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/weagings/"+weaging.ID.String(), nil)
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusNotFound {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusNotFound)
		}
	})

	t.Run("maps future events to a conflict", func(t *testing.T) {
		weaging := handlerWeaging()
		remove := &fakeDeleteWeagingUseCase{err: weagingdomain.ErrHasFutureEvents}
		handler := newTestHandlerWithDelete(&fakeCreateWeagingUseCase{}, &fakeUpdateWeagingUseCase{}, remove, &fakeListWeagingsUseCase{})
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/weagings/"+weaging.ID.String(), nil)
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusConflict {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusConflict)
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		weaging := handlerWeaging()
		remove := &fakeDeleteWeagingUseCase{err: errors.New("unexpected")}
		handler := newTestHandlerWithDelete(&fakeCreateWeagingUseCase{}, &fakeUpdateWeagingUseCase{}, remove, &fakeListWeagingsUseCase{})
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/weagings/"+weaging.ID.String(), nil)
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}
