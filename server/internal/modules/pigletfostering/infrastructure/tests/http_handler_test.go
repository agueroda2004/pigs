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
	pigletfosteringinfra "server/internal/modules/pigletfostering/infrastructure"
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
		if response.Body.Len() != 0 {
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
			{"date before partial weaging", pigletfosteringdomain.ErrMovementDateBeforePartialWeaging, http.StatusBadRequest},
			{"date in future", pigletfosteringdomain.ErrMovementDateInFuture, http.StatusBadRequest},
			{"invalid quantity", pigletfosteringdomain.ErrInvalidQuantity, http.StatusBadRequest},
			{"quantity above max", pigletfosteringdomain.ErrQuantityExceedsMax, http.StatusBadRequest},
			{"empty update", pigletfosteringdomain.ErrInvalidUpdate, http.StatusBadRequest},
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

func TestPigletFosteringHandlerUpdate(t *testing.T) {
	actorID := uuid.New()
	fosteringID := uuid.New()
	body := `{"movement_date":"2026-04-26","quantity":3,"note":"ajuste"}`

	t.Run("updates a fostering with the actor from context", func(t *testing.T) {
		update := &fakeUpdatePigletFosteringUseCase{fostering: handlerPigletFostering()}
		handler := newTestHandlerWithUpdate(&fakeCreatePigletFosteringUseCase{}, update, &fakeListPigletFosteringsUseCase{})
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/piglet-fosterings/"+fosteringID.String(), strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || !update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
		if update.id != fosteringID || update.command.UpdatedBy != actorID {
			t.Fatalf("id=%v command=%#v", update.id, update.command)
		}
		if update.command.Params.Quantity == nil || *update.command.Params.Quantity != 3 {
			t.Fatalf("unexpected quantity: %#v", update.command.Params.Quantity)
		}
		if !update.command.Params.MovementDate.Equal(time.Date(2026, time.April, 26, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected date: %v", update.command.Params.MovementDate)
		}
		if update.command.Params.Note == nil || *update.command.Params.Note != "ajuste" {
			t.Fatalf("unexpected note: %#v", update.command.Params.Note)
		}
		if response.Body.Len() != 0 {
			t.Fatalf("unexpected body: %s", response.Body.String())
		}
	})

	t.Run("accepts an update with only one field", func(t *testing.T) {
		update := &fakeUpdatePigletFosteringUseCase{fostering: handlerPigletFostering()}
		handler := newTestHandlerWithUpdate(&fakeCreatePigletFosteringUseCase{}, update, &fakeListPigletFosteringsUseCase{})
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/piglet-fosterings/"+fosteringID.String(), strings.NewReader(`{"quantity":4}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || !update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
		if update.command.Params.MovementDate != nil || update.command.Params.Note != nil {
			t.Fatalf("unexpected params: %#v", update.command.Params)
		}
	})

	t.Run("blocks update for non-admins", func(t *testing.T) {
		handler := newAdminGuardHandler(&fakeCreatePigletFosteringUseCase{})
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/piglet-fosterings/"+fosteringID.String(), strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusForbidden {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusForbidden)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		update := &fakeUpdatePigletFosteringUseCase{fostering: handlerPigletFostering()}
		handler := newTestHandlerWithUpdate(&fakeCreatePigletFosteringUseCase{}, update, &fakeListPigletFosteringsUseCase{})
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/piglet-fosterings/"+fosteringID.String(), strings.NewReader(body))
		response := serve(handler, request)

		if response.Code != http.StatusUnauthorized || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("rejects an invalid id", func(t *testing.T) {
		update := &fakeUpdatePigletFosteringUseCase{fostering: handlerPigletFostering()}
		handler := newTestHandlerWithUpdate(&fakeCreatePigletFosteringUseCase{}, update, &fakeListPigletFosteringsUseCase{})
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/piglet-fosterings/not-a-uuid", strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("rejects a change to the sows", func(t *testing.T) {
		update := &fakeUpdatePigletFosteringUseCase{fostering: handlerPigletFostering()}
		handler := newTestHandlerWithUpdate(&fakeCreatePigletFosteringUseCase{}, update, &fakeListPigletFosteringsUseCase{})
		invalid := fmt.Sprintf(`{"movement_date":"2026-04-26","quantity":3,"donor_sow_id":"%s"}`, uuid.New())
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/piglet-fosterings/"+fosteringID.String(), strings.NewReader(invalid))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("rejects an invalid date", func(t *testing.T) {
		update := &fakeUpdatePigletFosteringUseCase{fostering: handlerPigletFostering()}
		handler := newTestHandlerWithUpdate(&fakeCreatePigletFosteringUseCase{}, update, &fakeListPigletFosteringsUseCase{})
		request := httptest.NewRequest(http.MethodPatch, "/api/v1/piglet-fosterings/"+fosteringID.String(), strings.NewReader(`{"movement_date":"not-a-date"}`))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("maps application and domain errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"fostering not found", ports.ErrPigletFosteringNotFound, http.StatusNotFound},
			{"farrowing not found", ports.ErrFarrowingNotFound, http.StatusNotFound},
			{"insufficient piglets", farrowingdomain.ErrInsufficientPiglets, http.StatusConflict},
			{"empty update", pigletfosteringdomain.ErrInvalidUpdate, http.StatusBadRequest},
			{"quantity above max", pigletfosteringdomain.ErrQuantityExceedsMax, http.StatusBadRequest},
			{"date before partial weaging", pigletfosteringdomain.ErrMovementDateBeforePartialWeaging, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				update := &fakeUpdatePigletFosteringUseCase{err: test.err}
				handler := newTestHandlerWithUpdate(&fakeCreatePigletFosteringUseCase{}, update, &fakeListPigletFosteringsUseCase{})
				request := httptest.NewRequest(http.MethodPatch, "/api/v1/piglet-fosterings/"+fosteringID.String(), strings.NewReader(body))
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

func TestPigletFosteringHandlerDelete(t *testing.T) {
	fosteringID := uuid.New()
	actorID := uuid.New()
	target := "/api/v1/piglet-fosterings/" + fosteringID.String()

	deleteHandler := func(remove pigletfosteringinfra.DeletePigletFosteringUseCase) *pigletfosteringinfra.PigletFosteringHandler {
		return newTestHandlerWithDelete(
			&fakeCreatePigletFosteringUseCase{},
			&fakeUpdatePigletFosteringUseCase{},
			remove,
			&fakeListPigletFosteringsUseCase{},
		)
	}

	t.Run("deletes a fostering with the actor from context", func(t *testing.T) {
		remove := &fakeDeletePigletFosteringUseCase{}
		handler := deleteHandler(remove)
		request := authenticatedRequest(httptest.NewRequest(http.MethodDelete, target, nil), actorID)
		response := serve(handler, request)

		if response.Code != http.StatusNoContent || !remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
		if remove.id != fosteringID || remove.command.DeletedBy != actorID {
			t.Fatalf("unexpected command: %#v", remove.command)
		}
		if response.Body.Len() != 0 {
			t.Fatalf("unexpected body: %s", response.Body.String())
		}
	})

	t.Run("rejects an invalid id", func(t *testing.T) {
		remove := &fakeDeletePigletFosteringUseCase{}
		handler := deleteHandler(remove)
		request := authenticatedRequest(
			httptest.NewRequest(http.MethodDelete, "/api/v1/piglet-fosterings/not-a-uuid", nil), actorID,
		)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		remove := &fakeDeletePigletFosteringUseCase{}
		handler := deleteHandler(remove)
		response := serve(handler, httptest.NewRequest(http.MethodDelete, target, nil))

		if response.Code != http.StatusUnauthorized || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("blocks the deletion for non-admins", func(t *testing.T) {
		remove := &fakeDeletePigletFosteringUseCase{}
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
			{"fostering not found", ports.ErrPigletFosteringNotFound, http.StatusNotFound},
			{"farrowing not found", ports.ErrFarrowingNotFound, http.StatusNotFound},
			{"later weaging", pigletfosteringdomain.ErrHasLaterWeaging, http.StatusConflict},
			{"later partial weaging", pigletfosteringdomain.ErrHasLaterPartialWeaging, http.StatusConflict},
			{"insufficient piglets", farrowingdomain.ErrInsufficientPiglets, http.StatusConflict},
			{"missing actor", pigletfosteringdomain.ErrInvalidDeletedBy, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				remove := &fakeDeletePigletFosteringUseCase{err: test.err}
				handler := deleteHandler(remove)
				request := authenticatedRequest(httptest.NewRequest(http.MethodDelete, target, nil), actorID)
				response := serve(handler, request)

				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}
