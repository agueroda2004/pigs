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
	pigletdeathapplication "server/internal/modules/pigletdeath/application"
	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	"server/internal/modules/pigletdeath/ports"
)

func TestPigletDeathHandlerCreate(t *testing.T) {
	death := handlerPigletDeath()
	actorID := uuid.New()
	sowID := uuid.New()
	operatorID := uuid.New()

	body := fmt.Sprintf(
		`{"sow_id":"%s","operator_id":"%s","death_date":"2026-04-25",`+
			`"quantity":2,"weight":2.5,"cause":"Aplastado","turn":"Mañana","note":"muerte"}`,
		sowID, operatorID,
	)

	t.Run("creates a piglet death with the actor from context", func(t *testing.T) {
		create := &fakeCreatePigletDeathUseCase{death: death}
		handler := newTestHandler(create, &fakeListPigletDeathsUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/piglet-deaths", strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusCreated || create.command.CreatedBy != actorID {
			t.Fatalf("status=%d command=%#v", response.Code, create.command)
		}
		if create.command.SowID != sowID || create.command.OperatorID != operatorID {
			t.Fatalf("unexpected command: %#v", create.command)
		}
		if create.command.Quantity != 2 || create.command.Cause != pigletdeathdomain.CauseCrushed {
			t.Fatalf("unexpected command: %#v", create.command)
		}
		if create.command.Weight == nil || *create.command.Weight != 2.5 {
			t.Fatalf("unexpected weight: %#v", create.command.Weight)
		}
		if create.command.Turn != pigletdeathdomain.TurnMorning {
			t.Fatalf("unexpected turn: %#v", create.command.Turn)
		}
		if !create.command.DeathDate.Equal(time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected date: %v", create.command.DeathDate)
		}
		if response.Body.Len() != 0 {
			t.Fatalf("unexpected body: %s", response.Body.String())
		}
	})

	t.Run("blocks creation for non-admins", func(t *testing.T) {
		create := &fakeCreatePigletDeathUseCase{death: death}
		handler := newAdminGuardHandler(create)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/piglet-deaths", strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusForbidden || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		create := &fakeCreatePigletDeathUseCase{death: death}
		handler := newTestHandler(create, &fakeListPigletDeathsUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/piglet-deaths", strings.NewReader(body)))

		if response.Code != http.StatusUnauthorized || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("rejects an unknown field", func(t *testing.T) {
		create := &fakeCreatePigletDeathUseCase{death: death}
		handler := newTestHandler(create, &fakeListPigletDeathsUseCase{})
		invalid := fmt.Sprintf(
			`{"sow_id":"%s","operator_id":"%s","death_date":"2026-04-25","quantity":2,"cause":"Aplastado","turn":"Mañana","extra":true}`,
			sowID, operatorID,
		)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/piglet-deaths", strings.NewReader(invalid))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid fields", func(t *testing.T) {
		create := &fakeCreatePigletDeathUseCase{death: death}
		handler := newTestHandler(create, &fakeListPigletDeathsUseCase{})

		for _, invalid := range []string{
			`{"sow_id":"not-a-uuid","operator_id":"` + operatorID.String() + `","death_date":"2026-04-25","quantity":2,"cause":"Aplastado","turn":"Mañana"}`,
			fmt.Sprintf(`{"sow_id":"%s","operator_id":"not-a-uuid","death_date":"2026-04-25","quantity":2,"cause":"Aplastado","turn":"Mañana"}`, sowID),
			fmt.Sprintf(`{"sow_id":"%s","operator_id":"%s","death_date":"not-a-date","quantity":2,"cause":"Aplastado","turn":"Mañana"}`, sowID, operatorID),
			fmt.Sprintf(`{"sow_id":"%s","operator_id":"%s","death_date":"2026-04-25","quantity":2,"cause":"Mágica","turn":"Mañana"}`, sowID, operatorID),
			fmt.Sprintf(`{"sow_id":"%s","operator_id":"%s","death_date":"2026-04-25","quantity":2,"cause":"Aplastado","turn":"Noche"}`, sowID, operatorID),
		} {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/piglet-deaths", strings.NewReader(invalid))
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
			{"death not found", ports.ErrPigletDeathNotFound, http.StatusNotFound},
			{"sow not found", ports.ErrSowNotFound, http.StatusNotFound},
			{"farrowing not found", ports.ErrFarrowingNotFound, http.StatusNotFound},
			{"operator not found", ports.ErrOperatorNotFound, http.StatusNotFound},
			{"sow not lactating", pigletdeathapplication.ErrSowNotLactating, http.StatusConflict},
			{"insufficient piglets", farrowingdomain.ErrInsufficientPiglets, http.StatusConflict},
			{"date before farrowing", pigletdeathdomain.ErrDeathDateBeforeFarrowing, http.StatusBadRequest},
			{"date in future", pigletdeathdomain.ErrDeathDateInFuture, http.StatusBadRequest},
			{"date before partial weaging", pigletdeathdomain.ErrDeathDateBeforeWeaging, http.StatusBadRequest},
			{"invalid quantity", pigletdeathdomain.ErrInvalidQuantity, http.StatusBadRequest},
			{"quantity too high", pigletdeathdomain.ErrQuantityTooHigh, http.StatusBadRequest},
			{"weight required", pigletdeathdomain.ErrWeightRequired, http.StatusBadRequest},
			{"invalid weight", pigletdeathdomain.ErrInvalidWeight, http.StatusBadRequest},
			{"weight too high", pigletdeathdomain.ErrWeightTooHigh, http.StatusBadRequest},
			{"invalid cause", pigletdeathdomain.ErrInvalidCause, http.StatusBadRequest},
			{"invalid turn", pigletdeathdomain.ErrInvalidTurn, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				create := &fakeCreatePigletDeathUseCase{err: test.err}
				handler := newTestHandler(create, &fakeListPigletDeathsUseCase{})
				request := httptest.NewRequest(http.MethodPost, "/api/v1/piglet-deaths", strings.NewReader(body))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestPigletDeathHandlerList(t *testing.T) {
	t.Run("returns one page with its metadata and no audit fields", func(t *testing.T) {
		list := &fakeListPigletDeathsUseCase{
			deaths: []*pigletdeathdomain.PigletDeath{handlerPigletDeath()},
			total:  15,
		}
		handler := newTestHandler(&fakeCreatePigletDeathUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/piglet-deaths", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		body := response.Body.String()
		for _, want := range []string{
			`"items"`, "Aplastado", "2026-04-25", "sow_code", "C-001", "farrowing_id",
			"operator_name", "weight", `"total":15`, `"page":1`, `"page_size":10`, `"total_pages":2`,
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("body missing %q: %s", want, body)
			}
		}
		for _, unwanted := range []string{"created_at", "updated_at", "created_by", "updated_by"} {
			if strings.Contains(body, unwanted) {
				t.Fatalf("body should not contain %q: %s", unwanted, body)
			}
		}
	})

	t.Run("returns an empty items array when there are no deaths", func(t *testing.T) {
		list := &fakeListPigletDeathsUseCase{deaths: []*pigletdeathdomain.PigletDeath{}}
		handler := newTestHandler(&fakeCreatePigletDeathUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/piglet-deaths", nil))

		body := response.Body.String()
		if response.Code != http.StatusOK || !strings.Contains(body, `"items":[]`) {
			t.Fatalf("status=%d body=%s", response.Code, body)
		}
	})

	t.Run("parses the page and the sow code and date range filters", func(t *testing.T) {
		list := &fakeListPigletDeathsUseCase{}
		handler := newTestHandler(&fakeCreatePigletDeathUseCase{}, list)
		response := serve(handler, httptest.NewRequest(
			http.MethodGet,
			"/api/v1/piglet-deaths?sow_code=C-001&from=2026-04-01&to=2026-04-30&page=2",
			nil,
		))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		if list.page != 2 {
			t.Fatalf("page = %d, want 2", list.page)
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

	t.Run("returns an empty filter and page one when no params are given", func(t *testing.T) {
		list := &fakeListPigletDeathsUseCase{}
		handler := newTestHandler(&fakeCreatePigletDeathUseCase{}, list)
		serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/piglet-deaths", nil))

		if list.filter.SowCode != nil || list.filter.FromDate != nil || list.filter.ToDate != nil {
			t.Fatalf("unexpected filter: %#v", list.filter)
		}
		if list.page != 1 {
			t.Fatalf("page = %d, want 1", list.page)
		}
	})

	t.Run("rejects invalid filter and page values", func(t *testing.T) {
		list := &fakeListPigletDeathsUseCase{}
		handler := newTestHandler(&fakeCreatePigletDeathUseCase{}, list)

		for _, query := range []string{
			"sow_code=" + strings.Repeat("a", 51),
			"from=not-a-date",
			"to=not-a-date",
			"from=2026-04-30&to=2026-04-01",
			"page=0",
			"page=not-a-number",
		} {
			response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/piglet-deaths?"+query, nil))
			if response.Code != http.StatusBadRequest || list.called {
				t.Fatalf("status=%d called=%v query=%s", response.Code, list.called, query)
			}
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		list := &fakeListPigletDeathsUseCase{err: errors.New("unexpected")}
		handler := newTestHandler(&fakeCreatePigletDeathUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/piglet-deaths", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}

func TestPigletDeathHandlerDelete(t *testing.T) {
	deathID := uuid.New()
	actorID := uuid.New()
	target := "/api/v1/piglet-deaths/" + deathID.String()

	t.Run("deletes the death with the actor from context", func(t *testing.T) {
		remove := &fakeDeletePigletDeathUseCase{}
		handler := newDeleteHandler(remove)
		request := authenticatedRequest(httptest.NewRequest(http.MethodDelete, target, nil), actorID)
		response := serve(handler, request)

		if response.Code != http.StatusNoContent || !remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
		if remove.command.ID != deathID || remove.command.DeletedBy != actorID {
			t.Fatalf("unexpected command: %#v", remove.command)
		}
		if response.Body.Len() != 0 {
			t.Fatalf("unexpected body: %s", response.Body.String())
		}
	})

	t.Run("rejects an invalid id", func(t *testing.T) {
		remove := &fakeDeletePigletDeathUseCase{}
		handler := newDeleteHandler(remove)
		request := authenticatedRequest(
			httptest.NewRequest(http.MethodDelete, "/api/v1/piglet-deaths/not-a-uuid", nil), actorID,
		)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		remove := &fakeDeletePigletDeathUseCase{}
		handler := newDeleteHandler(remove)
		response := serve(handler, httptest.NewRequest(http.MethodDelete, target, nil))

		if response.Code != http.StatusUnauthorized || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("blocks the deletion for non-admins", func(t *testing.T) {
		remove := &fakeDeletePigletDeathUseCase{}
		handler := newAdminGuardDeleteHandler(remove)
		request := authenticatedRequest(httptest.NewRequest(http.MethodDelete, target, nil), actorID)
		response := serve(handler, request)

		if response.Code != http.StatusForbidden || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("maps the delete errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"death not found", ports.ErrPigletDeathNotFound, http.StatusNotFound},
			{"farrowing not found", ports.ErrFarrowingNotFound, http.StatusNotFound},
			{"later weaging", pigletdeathdomain.ErrHasLaterWeaging, http.StatusConflict},
			{"later partial weaging", pigletdeathdomain.ErrHasLaterPartialWeaging, http.StatusConflict},
			{"insufficient piglets", farrowingdomain.ErrInsufficientPiglets, http.StatusConflict},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				remove := &fakeDeletePigletDeathUseCase{err: test.err}
				handler := newDeleteHandler(remove)
				request := authenticatedRequest(httptest.NewRequest(http.MethodDelete, target, nil), actorID)
				response := serve(handler, request)

				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestPigletDeathHandlerUpdate(t *testing.T) {
	deathID := uuid.New()
	actorID := uuid.New()
	operatorID := uuid.New()
	target := "/api/v1/piglet-deaths/" + deathID.String()

	body := fmt.Sprintf(
		`{"operator_id":"%s","weight":3.5,"cause":"Bacteria","turn":"Tarde","note":"nueva nota"}`,
		operatorID,
	)

	t.Run("updates the death with the actor from context", func(t *testing.T) {
		update := &fakeUpdatePigletDeathUseCase{}
		handler := newUpdateHandler(update)
		request := authenticatedRequest(httptest.NewRequest(http.MethodPatch, target, strings.NewReader(body)), actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || !update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
		if update.id != deathID || update.command.UpdatedBy != actorID {
			t.Fatalf("unexpected command: id=%v %#v", update.id, update.command)
		}
		params := update.command.Params
		if params.OperatorID == nil || *params.OperatorID != operatorID {
			t.Fatalf("unexpected operator: %#v", params.OperatorID)
		}
		if params.Weight == nil || *params.Weight != 3.5 {
			t.Fatalf("unexpected weight: %#v", params.Weight)
		}
		if params.Cause == nil || *params.Cause != pigletdeathdomain.CauseBacteria {
			t.Fatalf("unexpected cause: %#v", params.Cause)
		}
		if params.Turn == nil || *params.Turn != pigletdeathdomain.TurnAfternoon {
			t.Fatalf("unexpected turn: %#v", params.Turn)
		}
		if params.Note == nil || *params.Note != "nueva nota" {
			t.Fatalf("unexpected note: %#v", params.Note)
		}
		if response.Body.Len() != 0 {
			t.Fatalf("unexpected body: %s", response.Body.String())
		}
	})

	t.Run("parses the locked fields when they are sent", func(t *testing.T) {
		update := &fakeUpdatePigletDeathUseCase{}
		handler := newUpdateHandler(update)
		sowID := uuid.New()
		locked := fmt.Sprintf(
			`{"death_date":"2026-04-25","quantity":2,"sow_id":"%s","weight":3.5}`,
			sowID,
		)
		request := authenticatedRequest(httptest.NewRequest(http.MethodPatch, target, strings.NewReader(locked)), actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusOK)
		}
		params := update.command.Params
		if params.DeathDate == nil || !params.DeathDate.Equal(time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected death date: %#v", params.DeathDate)
		}
		if params.Quantity == nil || *params.Quantity != 2 {
			t.Fatalf("unexpected quantity: %#v", params.Quantity)
		}
		if params.SowID == nil || *params.SowID != sowID {
			t.Fatalf("unexpected sow: %#v", params.SowID)
		}
	})

	t.Run("rejects an invalid id", func(t *testing.T) {
		update := &fakeUpdatePigletDeathUseCase{}
		handler := newUpdateHandler(update)
		request := authenticatedRequest(
			httptest.NewRequest(http.MethodPatch, "/api/v1/piglet-deaths/not-a-uuid", strings.NewReader(body)),
			actorID,
		)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("rejects invalid fields", func(t *testing.T) {
		update := &fakeUpdatePigletDeathUseCase{}
		handler := newUpdateHandler(update)

		for _, invalid := range []string{
			`{"death_date":"not-a-date","weight":3.5}`,
			`{"sow_id":"not-a-uuid","weight":3.5}`,
			`{"operator_id":"not-a-uuid","weight":3.5}`,
			`{"cause":"Mágica","weight":3.5}`,
			`{"turn":"Noche","weight":3.5}`,
			`{"weight":3.5,"extra":true}`,
		} {
			request := authenticatedRequest(
				httptest.NewRequest(http.MethodPatch, target, strings.NewReader(invalid)), actorID,
			)
			response := serve(handler, request)

			if response.Code != http.StatusBadRequest || update.called {
				t.Fatalf("status=%d called=%v body=%s", response.Code, update.called, invalid)
			}
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		update := &fakeUpdatePigletDeathUseCase{}
		handler := newUpdateHandler(update)
		response := serve(handler, httptest.NewRequest(http.MethodPatch, target, strings.NewReader(body)))

		if response.Code != http.StatusUnauthorized || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("blocks the update for non-admins", func(t *testing.T) {
		update := &fakeUpdatePigletDeathUseCase{}
		handler := newAdminGuardUpdateHandler(update)
		request := authenticatedRequest(httptest.NewRequest(http.MethodPatch, target, strings.NewReader(body)), actorID)
		response := serve(handler, request)

		if response.Code != http.StatusForbidden || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("maps the update errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"death not found", ports.ErrPigletDeathNotFound, http.StatusNotFound},
			{"operator not found", ports.ErrOperatorNotFound, http.StatusNotFound},
			{"immutable field", pigletdeathdomain.ErrImmutableDeathField, http.StatusBadRequest},
			{"invalid update", pigletdeathdomain.ErrInvalidUpdate, http.StatusBadRequest},
			{"weight required", pigletdeathdomain.ErrWeightRequired, http.StatusBadRequest},
			{"weight too high", pigletdeathdomain.ErrWeightTooHigh, http.StatusBadRequest},
			{"invalid cause", pigletdeathdomain.ErrInvalidCause, http.StatusBadRequest},
			{"invalid note", pigletdeathdomain.ErrInvalidNote, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				update := &fakeUpdatePigletDeathUseCase{err: test.err}
				handler := newUpdateHandler(update)
				request := authenticatedRequest(
					httptest.NewRequest(http.MethodPatch, target, strings.NewReader(body)), actorID,
				)
				response := serve(handler, request)

				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}
