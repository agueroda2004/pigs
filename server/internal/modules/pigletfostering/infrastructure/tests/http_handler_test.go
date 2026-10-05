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
	pigletfosteringapplication "server/internal/modules/pigletfostering/application"
	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
	"server/internal/modules/pigletfostering/ports"
)

func TestPigletFosteringHandlerCreate(t *testing.T) {
	fostering := handlerPigletFostering()
	actorID := uuid.New()
	donorSowID := uuid.New()
	receiverSowID := uuid.New()

	body := fmt.Sprintf(
		`{"donor_sow_id":"%s","receiver_sow_id":"%s","movement_date":"2026-04-25",`+
			`"quantity":2,"note":"traslado"}`,
		donorSowID, receiverSowID,
	)

	t.Run("creates a fostering with the actor from context", func(t *testing.T) {
		create := &fakeCreatePigletFosteringUseCase{fostering: fostering}
		handler := newTestHandler(create, &fakeListPigletFosteringsUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/piglet-fosterings", strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusCreated || create.command.CreatedBy != actorID {
			t.Fatalf("status=%d command=%#v", response.Code, create.command)
		}
		if create.command.DonorSowID != donorSowID || create.command.ReceiverSowID != receiverSowID {
			t.Fatalf("unexpected command: %#v", create.command)
		}
		if create.command.Quantity != 2 {
			t.Fatalf("unexpected quantity: %#v", create.command.Quantity)
		}
		if !create.command.MovementDate.Equal(time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected date: %v", create.command.MovementDate)
		}
		if create.command.Note == nil || *create.command.Note != "traslado" {
			t.Fatalf("unexpected note: %#v", create.command.Note)
		}
		if !strings.Contains(response.Body.String(), "2026-04-25") {
			t.Fatalf("unexpected body: %s", response.Body.String())
		}
	})

	t.Run("blocks creation for non-admins", func(t *testing.T) {
		create := &fakeCreatePigletFosteringUseCase{fostering: fostering}
		handler := newAdminGuardHandler(create)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/piglet-fosterings", strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusForbidden || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		create := &fakeCreatePigletFosteringUseCase{fostering: fostering}
		handler := newTestHandler(create, &fakeListPigletFosteringsUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/piglet-fosterings", strings.NewReader(body)))

		if response.Code != http.StatusUnauthorized || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("rejects an unknown field", func(t *testing.T) {
		create := &fakeCreatePigletFosteringUseCase{fostering: fostering}
		handler := newTestHandler(create, &fakeListPigletFosteringsUseCase{})
		invalid := fmt.Sprintf(
			`{"donor_sow_id":"%s","receiver_sow_id":"%s","movement_date":"2026-04-25","quantity":2,"extra":true}`,
			donorSowID, receiverSowID,
		)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/piglet-fosterings", strings.NewReader(invalid))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns bad request for invalid fields", func(t *testing.T) {
		create := &fakeCreatePigletFosteringUseCase{fostering: fostering}
		handler := newTestHandler(create, &fakeListPigletFosteringsUseCase{})

		for _, invalid := range []string{
			`{"donor_sow_id":"not-a-uuid","receiver_sow_id":"` + receiverSowID.String() + `","movement_date":"2026-04-25","quantity":2}`,
			fmt.Sprintf(`{"donor_sow_id":"%s","receiver_sow_id":"not-a-uuid","movement_date":"2026-04-25","quantity":2}`, donorSowID),
			fmt.Sprintf(`{"donor_sow_id":"%s","receiver_sow_id":"%s","movement_date":"not-a-date","quantity":2}`, donorSowID, receiverSowID),
		} {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/piglet-fosterings", strings.NewReader(invalid))
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
			{"fostering not found", ports.ErrPigletFosteringNotFound, http.StatusNotFound},
			{"sow not found", ports.ErrSowNotFound, http.StatusNotFound},
			{"farrowing not found", ports.ErrFarrowingNotFound, http.StatusNotFound},
			{"sow not lactating", pigletfosteringapplication.ErrSowNotLactating, http.StatusConflict},
			{"insufficient piglets", farrowingdomain.ErrInsufficientPiglets, http.StatusConflict},
			{"same farrowing", pigletfosteringdomain.ErrSameFarrowing, http.StatusBadRequest},
			{"date before farrowing", pigletfosteringdomain.ErrMovementDateBeforeFarrowing, http.StatusBadRequest},
			{"date in future", pigletfosteringdomain.ErrMovementDateInFuture, http.StatusBadRequest},
			{"invalid quantity", pigletfosteringdomain.ErrInvalidQuantity, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				create := &fakeCreatePigletFosteringUseCase{err: test.err}
				handler := newTestHandler(create, &fakeListPigletFosteringsUseCase{})
				request := httptest.NewRequest(http.MethodPost, "/api/v1/piglet-fosterings", strings.NewReader(body))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestPigletFosteringHandlerList(t *testing.T) {
	t.Run("returns every fostering", func(t *testing.T) {
		list := &fakeListPigletFosteringsUseCase{fosterings: []*pigletfosteringdomain.PigletFostering{handlerPigletFostering()}}
		handler := newTestHandler(&fakeCreatePigletFosteringUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/piglet-fosterings", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		body := response.Body.String()
		for _, want := range []string{"2026-04-25", "donor_farrowing_id", "receiver_farrowing_id", "donor_sow_id", "receiver_sow_id", "quantity"} {
			if !strings.Contains(body, want) {
				t.Fatalf("body missing %q: %s", want, body)
			}
		}
	})

	t.Run("returns an empty array when there are no fosterings", func(t *testing.T) {
		list := &fakeListPigletFosteringsUseCase{fosterings: []*pigletfosteringdomain.PigletFostering{}}
		handler := newTestHandler(&fakeCreatePigletFosteringUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/piglet-fosterings", nil))

		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("parses the sow and date range filters", func(t *testing.T) {
		donorSowID := uuid.New()
		receiverSowID := uuid.New()
		list := &fakeListPigletFosteringsUseCase{}
		handler := newTestHandler(&fakeCreatePigletFosteringUseCase{}, list)
		response := serve(handler, httptest.NewRequest(
			http.MethodGet,
			"/api/v1/piglet-fosterings?donor_sow_id="+donorSowID.String()+"&receiver_sow_id="+receiverSowID.String()+"&from=2026-04-01&to=2026-04-30",
			nil,
		))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		if list.filter.DonorSowID == nil || *list.filter.DonorSowID != donorSowID {
			t.Fatalf("unexpected donor sow id: %#v", list.filter.DonorSowID)
		}
		if list.filter.ReceiverSowID == nil || *list.filter.ReceiverSowID != receiverSowID {
			t.Fatalf("unexpected receiver sow id: %#v", list.filter.ReceiverSowID)
		}
		if list.filter.FromDate == nil || !list.filter.FromDate.Equal(time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected from date: %#v", list.filter.FromDate)
		}
		if list.filter.ToDate == nil || !list.filter.ToDate.Equal(time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected to date: %#v", list.filter.ToDate)
		}
	})

	t.Run("returns an empty filter when no params are given", func(t *testing.T) {
		list := &fakeListPigletFosteringsUseCase{}
		handler := newTestHandler(&fakeCreatePigletFosteringUseCase{}, list)
		serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/piglet-fosterings", nil))

		if list.filter.DonorSowID != nil || list.filter.ReceiverSowID != nil || list.filter.FromDate != nil || list.filter.ToDate != nil {
			t.Fatalf("unexpected filter: %#v", list.filter)
		}
	})

	t.Run("rejects invalid filter values", func(t *testing.T) {
		list := &fakeListPigletFosteringsUseCase{}
		handler := newTestHandler(&fakeCreatePigletFosteringUseCase{}, list)

		for _, query := range []string{
			"donor_sow_id=not-a-uuid",
			"receiver_sow_id=not-a-uuid",
			"from=not-a-date",
			"to=not-a-date",
			"from=2026-04-30&to=2026-04-01",
		} {
			response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/piglet-fosterings?"+query, nil))
			if response.Code != http.StatusBadRequest || list.called {
				t.Fatalf("status=%d called=%v query=%s", response.Code, list.called, query)
			}
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		list := &fakeListPigletFosteringsUseCase{err: errors.New("unexpected")}
		handler := newTestHandler(&fakeCreatePigletFosteringUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/piglet-fosterings", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}
