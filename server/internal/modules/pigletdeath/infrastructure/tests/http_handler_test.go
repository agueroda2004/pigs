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
		if !strings.Contains(response.Body.String(), "Aplastado") {
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
			{"invalid quantity", pigletdeathdomain.ErrInvalidQuantity, http.StatusBadRequest},
			{"invalid weight", pigletdeathdomain.ErrInvalidWeight, http.StatusBadRequest},
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
	t.Run("returns every piglet death", func(t *testing.T) {
		list := &fakeListPigletDeathsUseCase{deaths: []*pigletdeathdomain.PigletDeath{handlerPigletDeath()}}
		handler := newTestHandler(&fakeCreatePigletDeathUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/piglet-deaths", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		body := response.Body.String()
		for _, want := range []string{"Aplastado", "2026-04-25", "sow_id", "farrowing_id", "operator_name", "weight"} {
			if !strings.Contains(body, want) {
				t.Fatalf("body missing %q: %s", want, body)
			}
		}
	})

	t.Run("returns an empty array when there are no deaths", func(t *testing.T) {
		list := &fakeListPigletDeathsUseCase{deaths: []*pigletdeathdomain.PigletDeath{}}
		handler := newTestHandler(&fakeCreatePigletDeathUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/piglet-deaths", nil))

		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("parses the sow and date range filters", func(t *testing.T) {
		sowID := uuid.New()
		list := &fakeListPigletDeathsUseCase{}
		handler := newTestHandler(&fakeCreatePigletDeathUseCase{}, list)
		response := serve(handler, httptest.NewRequest(
			http.MethodGet,
			"/api/v1/piglet-deaths?sow_id="+sowID.String()+"&from=2026-04-01&to=2026-04-30",
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
		list := &fakeListPigletDeathsUseCase{}
		handler := newTestHandler(&fakeCreatePigletDeathUseCase{}, list)
		serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/piglet-deaths", nil))

		if list.filter.SowID != nil || list.filter.FromDate != nil || list.filter.ToDate != nil {
			t.Fatalf("unexpected filter: %#v", list.filter)
		}
	})

	t.Run("rejects invalid filter values", func(t *testing.T) {
		list := &fakeListPigletDeathsUseCase{}
		handler := newTestHandler(&fakeCreatePigletDeathUseCase{}, list)

		for _, query := range []string{
			"sow_id=not-a-uuid",
			"from=not-a-date",
			"to=not-a-date",
			"from=2026-04-30&to=2026-04-01",
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
