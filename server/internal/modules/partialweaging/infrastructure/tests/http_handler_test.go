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
	partialweagingapplication "server/internal/modules/partialweaging/application"
	partialweagingdomain "server/internal/modules/partialweaging/domain"
	"server/internal/modules/partialweaging/ports"
)

func TestPartialWeagingHandlerCreate(t *testing.T) {
	weaging := handlerPartialWeaging()
	actorID := uuid.New()
	sowID := uuid.New()

	body := fmt.Sprintf(
		`{"sow_id":"%s","weaging_date":"2026-04-25","quantity":2,`+
			`"total_weight":42.5,"type":"Normal","note":"destete"}`,
		sowID,
	)

	t.Run("creates a partial weaging with the actor from context", func(t *testing.T) {
		create := &fakeCreatePartialWeagingUseCase{weaging: weaging}
		handler := newTestHandler(create, &fakeListPartialWeagingsUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/partial-weagings", strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusCreated || create.command.CreatedBy != actorID {
			t.Fatalf("status=%d command=%#v", response.Code, create.command)
		}
		if create.command.SowID != sowID {
			t.Fatalf("unexpected stall command: %#v", create.command)
		}
		if create.command.Quantity != 2 || create.command.Type != partialweagingdomain.TypeNormal {
			t.Fatalf("unexpected command: %#v", create.command)
		}
		if create.command.TotalWeight == nil || *create.command.TotalWeight != 42.5 {
			t.Fatalf("unexpected weight: %#v", create.command.TotalWeight)
		}
		if !create.command.WeagingDate.Equal(time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected date: %v", create.command.WeagingDate)
		}
		if strings.TrimSpace(response.Body.String()) != "" {
			t.Fatalf("body should be empty, got: %s", response.Body.String())
		}
	})

	t.Run("blocks creation for non-admins", func(t *testing.T) {
		create := &fakeCreatePartialWeagingUseCase{weaging: weaging}
		handler := newAdminGuardHandler(create)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/partial-weagings", strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusForbidden || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		create := &fakeCreatePartialWeagingUseCase{weaging: weaging}
		handler := newTestHandler(create, &fakeListPartialWeagingsUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/partial-weagings", strings.NewReader(body)))

		if response.Code != http.StatusUnauthorized || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("rejects an unknown field", func(t *testing.T) {
		create := &fakeCreatePartialWeagingUseCase{weaging: weaging}
		handler := newTestHandler(create, &fakeListPartialWeagingsUseCase{})
		invalid := fmt.Sprintf(
			`{"sow_id":"%s","weaging_date":"2026-04-25","quantity":2,"type":"Normal","extra":true}`,
			sowID,
		)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/partial-weagings", strings.NewReader(invalid))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid fields", func(t *testing.T) {
		create := &fakeCreatePartialWeagingUseCase{weaging: weaging}
		handler := newTestHandler(create, &fakeListPartialWeagingsUseCase{})

		for _, invalid := range []string{
			`{"sow_id":"not-a-uuid","weaging_date":"2026-04-25","quantity":2,"type":"Normal"}`,
			fmt.Sprintf(`{"sow_id":"%s","weaging_date":"not-a-date","quantity":2,"type":"Normal"}`, sowID),
			fmt.Sprintf(`{"sow_id":"%s","weaging_date":"2026-04-25","quantity":2,"type":"Nope"}`, sowID),
		} {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/partial-weagings", strings.NewReader(invalid))
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
			{"partial weaging not found", ports.ErrPartialWeagingNotFound, http.StatusNotFound},
			{"sow not found", ports.ErrSowNotFound, http.StatusNotFound},
			{"farrowing not found", ports.ErrFarrowingNotFound, http.StatusNotFound},
			{"sow not lactating", partialweagingapplication.ErrSowNotLactating, http.StatusConflict},
			{"insufficient piglets", farrowingdomain.ErrInsufficientPiglets, http.StatusConflict},
			{"invalid type", partialweagingdomain.ErrInvalidType, http.StatusBadRequest},
			{"date before farrowing", partialweagingdomain.ErrWeagingDateBeforeFarrowing, http.StatusBadRequest},
			{"date before events", partialweagingdomain.ErrWeagingDateBeforeEvents, http.StatusBadRequest},
			{"date in future", partialweagingdomain.ErrWeagingDateInFuture, http.StatusBadRequest},
			{"invalid quantity", partialweagingdomain.ErrInvalidQuantity, http.StatusBadRequest},
			{"quantity above maximum", partialweagingdomain.ErrQuantityExceedsMax, http.StatusBadRequest},
			{"quantity mismatch", partialweagingdomain.ErrQuantityMismatch, http.StatusBadRequest},
			{"weight above maximum", partialweagingdomain.ErrTotalWeightExceedsMax, http.StatusBadRequest},
			{"date before last partial weaging", partialweagingdomain.ErrWeagingDateBeforePartial, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				create := &fakeCreatePartialWeagingUseCase{err: test.err}
				handler := newTestHandler(create, &fakeListPartialWeagingsUseCase{})
				request := httptest.NewRequest(http.MethodPost, "/api/v1/partial-weagings", strings.NewReader(body))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestPartialWeagingHandlerUpdate(t *testing.T) {
	id := uuid.New()
	actorID := uuid.New()
	body := `{"total_weight":55.5,"note":"nota"}`

	t.Run("updates the partial weaging and returns OK", func(t *testing.T) {
		update := &fakeUpdatePartialWeagingUseCase{}
		handler := newUpdateTestHandler(update)
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/partial-weagings/"+id.String(), strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || !update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
		if update.id != id || update.command.UpdatedBy != actorID {
			t.Fatalf("unexpected command: %#v", update.command)
		}
		if update.command.TotalWeight == nil || *update.command.TotalWeight != 55.5 {
			t.Fatalf("unexpected weight: %#v", update.command.TotalWeight)
		}
		if update.command.Note == nil || *update.command.Note != "nota" {
			t.Fatalf("unexpected note: %#v", update.command.Note)
		}
		if strings.TrimSpace(response.Body.String()) != "" {
			t.Fatalf("body should be empty, got: %s", response.Body.String())
		}
	})

	t.Run("clears the weight", func(t *testing.T) {
		update := &fakeUpdatePartialWeagingUseCase{}
		handler := newUpdateTestHandler(update)
		request := authenticatedRequest(
			httptest.NewRequest(http.MethodPatch, "/api/v1/partial-weagings/"+id.String(), strings.NewReader(`{"clear_total_weight":true}`)),
			actorID,
		)
		response := serve(handler, request)

		if response.Code != http.StatusOK || !update.command.ClearTotalWeight || update.command.TotalWeight != nil {
			t.Fatalf("status=%d command=%#v", response.Code, update.command)
		}
	})

	t.Run("rejects an invalid identifier", func(t *testing.T) {
		update := &fakeUpdatePartialWeagingUseCase{}
		handler := newUpdateTestHandler(update)
		request := authenticatedRequest(
			httptest.NewRequest(http.MethodPatch, "/api/v1/partial-weagings/not-a-uuid", strings.NewReader(body)),
			actorID,
		)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("rejects immutable fields", func(t *testing.T) {
		for _, invalid := range []string{
			`{"quantity":2}`,
			`{"sow_id":"` + uuid.New().String() + `"}`,
			`{"type":"Normal"}`,
			`{"farrowing_id":"` + uuid.New().String() + `"}`,
			`{"weaging_date":"2026-04-27"}`,
		} {
			update := &fakeUpdatePartialWeagingUseCase{}
			handler := newUpdateTestHandler(update)
			request := authenticatedRequest(
				httptest.NewRequest(http.MethodPatch, "/api/v1/partial-weagings/"+id.String(), strings.NewReader(invalid)),
				actorID,
			)
			response := serve(handler, request)

			if response.Code != http.StatusBadRequest || update.called {
				t.Fatalf("status=%d called=%v body=%s", response.Code, update.called, invalid)
			}
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		update := &fakeUpdatePartialWeagingUseCase{}
		handler := newUpdateTestHandler(update)
		response := serve(handler, httptest.NewRequest(http.MethodPatch, "/api/v1/partial-weagings/"+id.String(), strings.NewReader(body)))

		if response.Code != http.StatusUnauthorized || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("blocks updates for non-admins", func(t *testing.T) {
		handler := newAdminGuardHandler(&fakeCreatePartialWeagingUseCase{})
		request := authenticatedRequest(
			httptest.NewRequest(http.MethodPatch, "/api/v1/partial-weagings/"+id.String(), strings.NewReader(body)),
			actorID,
		)
		response := serve(handler, request)

		if response.Code != http.StatusForbidden {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusForbidden)
		}
	})

	t.Run("maps application and domain errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"partial weaging not found", ports.ErrPartialWeagingNotFound, http.StatusNotFound},
			{"farrowing not found", ports.ErrFarrowingNotFound, http.StatusNotFound},
			{"sow not found", ports.ErrSowNotFound, http.StatusNotFound},
			{"invalid update", partialweagingdomain.ErrInvalidUpdate, http.StatusBadRequest},
			{"invalid updated by", partialweagingdomain.ErrInvalidUpdatedBy, http.StatusBadRequest},
			{"invalid weight", partialweagingdomain.ErrInvalidTotalWeight, http.StatusBadRequest},
			{"weight above maximum", partialweagingdomain.ErrTotalWeightExceedsMax, http.StatusBadRequest},
			{"invalid note", partialweagingdomain.ErrInvalidNote, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				handler := newUpdateTestHandler(&fakeUpdatePartialWeagingUseCase{err: test.err})
				request := authenticatedRequest(
					httptest.NewRequest(http.MethodPatch, "/api/v1/partial-weagings/"+id.String(), strings.NewReader(body)),
					actorID,
				)
				response := serve(handler, request)

				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestPartialWeagingHandlerDelete(t *testing.T) {
	id := uuid.New()
	actorID := uuid.New()

	t.Run("deletes the partial weaging and returns no content", func(t *testing.T) {
		deleteUseCase := &fakeDeletePartialWeagingUseCase{}
		handler := newDeleteTestHandler(deleteUseCase)
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/partial-weagings/"+id.String(), nil)
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusNoContent || !deleteUseCase.called {
			t.Fatalf("status=%d called=%v", response.Code, deleteUseCase.called)
		}
		if deleteUseCase.command.ID != id || deleteUseCase.command.DeletedBy != actorID {
			t.Fatalf("unexpected command: %#v", deleteUseCase.command)
		}
		if strings.TrimSpace(response.Body.String()) != "" {
			t.Fatalf("body should be empty, got: %s", response.Body.String())
		}
	})

	t.Run("rejects an invalid identifier", func(t *testing.T) {
		deleteUseCase := &fakeDeletePartialWeagingUseCase{}
		handler := newDeleteTestHandler(deleteUseCase)
		request := authenticatedRequest(
			httptest.NewRequest(http.MethodDelete, "/api/v1/partial-weagings/not-a-uuid", nil),
			actorID,
		)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || deleteUseCase.called {
			t.Fatalf("status=%d called=%v", response.Code, deleteUseCase.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		deleteUseCase := &fakeDeletePartialWeagingUseCase{}
		handler := newDeleteTestHandler(deleteUseCase)
		response := serve(handler, httptest.NewRequest(http.MethodDelete, "/api/v1/partial-weagings/"+id.String(), nil))

		if response.Code != http.StatusUnauthorized || deleteUseCase.called {
			t.Fatalf("status=%d called=%v", response.Code, deleteUseCase.called)
		}
	})

	t.Run("maps application and domain errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"partial weaging not found", ports.ErrPartialWeagingNotFound, http.StatusNotFound},
			{"farrowing not found", ports.ErrFarrowingNotFound, http.StatusNotFound},
			{"sow not found", ports.ErrSowNotFound, http.StatusNotFound},
			{"future events", partialweagingdomain.ErrHasFutureEvents, http.StatusConflict},
			{"later nurse events", partialweagingdomain.ErrHasLaterNurseEvents, http.StatusConflict},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				handler := newDeleteTestHandler(&fakeDeletePartialWeagingUseCase{err: test.err})
				request := authenticatedRequest(
					httptest.NewRequest(http.MethodDelete, "/api/v1/partial-weagings/"+id.String(), nil),
					actorID,
				)
				response := serve(handler, request)

				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestPartialWeagingHandlerList(t *testing.T) {
	t.Run("returns the first page with its pagination metadata", func(t *testing.T) {
		list := &fakeListPartialWeagingsUseCase{
			weagings: []*partialweagingdomain.PartialWeaging{handlerPartialWeaging()},
			total:    1,
		}
		handler := newTestHandler(&fakeCreatePartialWeagingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/partial-weagings", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		if list.page != 1 {
			t.Fatalf("page = %d, want 1", list.page)
		}
		body := response.Body.String()
		for _, want := range []string{
			`"items"`, `"total":1`, `"page":1`, `"page_size":10`, `"total_pages":1`,
			"2026-04-25", "farrowing_id", "sow_id", "sow_code", "C-001", "quantity",
			"total_weight", "type",
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("body missing %q: %s", want, body)
			}
		}
		for _, omitted := range []string{"created_at", "updated_at", "created_by", "updated_by"} {
			if strings.Contains(body, omitted) {
				t.Fatalf("body should omit %q: %s", omitted, body)
			}
		}
	})

	t.Run("returns an empty page when there are no partial weagings", func(t *testing.T) {
		list := &fakeListPartialWeagingsUseCase{weagings: []*partialweagingdomain.PartialWeaging{}}
		handler := newTestHandler(&fakeCreatePartialWeagingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/partial-weagings", nil))

		body := response.Body.String()
		if response.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", response.Code, body)
		}
		for _, want := range []string{`"items":[]`, `"total":0`, `"total_pages":0`} {
			if !strings.Contains(body, want) {
				t.Fatalf("body missing %q: %s", want, body)
			}
		}
	})

	t.Run("parses the sow code, date range and page filters", func(t *testing.T) {
		list := &fakeListPartialWeagingsUseCase{}
		handler := newTestHandler(&fakeCreatePartialWeagingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(
			http.MethodGet,
			"/api/v1/partial-weagings?sow_code=%20C-001%20&from=2026-04-01&to=2026-04-30&page=2",
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
		if list.page != 2 {
			t.Fatalf("page = %d, want 2", list.page)
		}
	})

	t.Run("returns an empty filter and the first page when no params are given", func(t *testing.T) {
		list := &fakeListPartialWeagingsUseCase{}
		handler := newTestHandler(&fakeCreatePartialWeagingUseCase{}, list)
		serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/partial-weagings", nil))

		if list.filter.SowCode != nil || list.filter.FromDate != nil || list.filter.ToDate != nil {
			t.Fatalf("unexpected filter: %#v", list.filter)
		}
		if list.page != 1 {
			t.Fatalf("page = %d, want 1", list.page)
		}
	})

	t.Run("rejects invalid filter values", func(t *testing.T) {
		list := &fakeListPartialWeagingsUseCase{}
		handler := newTestHandler(&fakeCreatePartialWeagingUseCase{}, list)

		for _, query := range []string{
			"sow_code=" + strings.Repeat("A", 51),
			"from=not-a-date",
			"to=not-a-date",
			"from=2026-04-30&to=2026-04-01",
		} {
			response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/partial-weagings?"+query, nil))
			if response.Code != http.StatusBadRequest || list.called {
				t.Fatalf("status=%d called=%v query=%s", response.Code, list.called, query)
			}
		}
	})

	t.Run("rejects invalid pages", func(t *testing.T) {
		list := &fakeListPartialWeagingsUseCase{}
		handler := newTestHandler(&fakeCreatePartialWeagingUseCase{}, list)

		for _, query := range []string{"page=0", "page=-1", "page=abc"} {
			response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/partial-weagings?"+query, nil))
			if response.Code != http.StatusBadRequest || list.called {
				t.Fatalf("status=%d called=%v query=%s", response.Code, list.called, query)
			}
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		list := &fakeListPartialWeagingsUseCase{err: errors.New("unexpected")}
		handler := newTestHandler(&fakeCreatePartialWeagingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/partial-weagings", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}
