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

	farrowingapplication "server/internal/modules/farrowing/application"
	farrowingdomain "server/internal/modules/farrowing/domain"
	"server/internal/modules/farrowing/ports"
)

func TestFarrowingHandlerCreate(t *testing.T) {
	farrowing := handlerFarrowing()
	actorID := uuid.New()
	sowID := uuid.New()
	operatorID := uuid.New()
	medicationID := uuid.New()

	body := fmt.Sprintf(
		`{"sow_id":"%s","farrow_date":"2026-04-20",`+
			`"start_time":"22:00","end_time":"02:00","location":"Corral 3",`+
			`"live_born":10,"stillborn":1,"mummified":0,"is_manipulated":true,`+
			`"litter_weight":15.5,"stillborn_weight":1.5,`+
			`"operators":[{"operator_id":"%s"}],`+
			`"medications":[{"medication_id":"%s","dose":2,"applied_by":"%s"}]}`,
		sowID, operatorID, medicationID, operatorID,
	)

	t.Run("creates a farrowing with the actor from context", func(t *testing.T) {
		create := &fakeCreateFarrowingUseCase{farrowing: farrowing}
		handler := newTestHandler(create, &fakeListFarrowingsUseCase{})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/farrowings", strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusCreated || create.command.CreatedBy != actorID {
			t.Fatalf("status=%d command=%#v", response.Code, create.command)
		}
		if create.command.SowID != sowID {
			t.Fatalf("unexpected command: %#v", create.command)
		}
		if create.command.LiveBorn != 10 || !create.command.IsManipulated {
			t.Fatalf("unexpected command: %#v", create.command)
		}
		if create.command.LitterWeight == nil || *create.command.LitterWeight != 15.5 {
			t.Fatalf("unexpected litter weight: %#v", create.command.LitterWeight)
		}
		if create.command.StillbornWeight == nil || *create.command.StillbornWeight != 1.5 {
			t.Fatalf("unexpected stillborn weight: %#v", create.command.StillbornWeight)
		}
		if !create.command.FarrowDate.Equal(time.Date(2026, time.April, 20, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected date: %v", create.command.FarrowDate)
		}
		if len(create.command.Operators) != 1 || create.command.Operators[0].OperatorID != operatorID {
			t.Fatalf("unexpected operators: %#v", create.command.Operators)
		}
		if len(create.command.Medications) != 1 || create.command.Medications[0].MedicationID != medicationID {
			t.Fatalf("unexpected medications: %#v", create.command.Medications)
		}
		if response.Body.Len() != 0 {
			t.Fatalf("expected an empty body, got %s", response.Body.String())
		}
	})

	t.Run("accepts a decimal medication dose", func(t *testing.T) {
		create := &fakeCreateFarrowingUseCase{farrowing: farrowing}
		handler := newTestHandler(create, &fakeListFarrowingsUseCase{})
		decimalBody := fmt.Sprintf(
			`{"sow_id":"%s","farrow_date":"2026-04-20",`+
				`"live_born":10,"stillborn":0,"mummified":0,`+
				`"litter_weight":15.5,"stillborn_weight":1.5,`+
				`"medications":[{"medication_id":"%s","dose":0.5,"applied_by":"%s"}]}`,
			sowID, medicationID, operatorID,
		)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/farrowings", strings.NewReader(decimalBody))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusCreated || !create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
		if len(create.command.Medications) != 1 || create.command.Medications[0].Dose != 0.5 {
			t.Fatalf("unexpected medications: %#v", create.command.Medications)
		}
	})

	t.Run("blocks creation for non-admins", func(t *testing.T) {
		create := &fakeCreateFarrowingUseCase{farrowing: farrowing}
		handler := newAdminGuardHandler(create)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/farrowings", strings.NewReader(body))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusForbidden || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("rejects an unknown field", func(t *testing.T) {
		create := &fakeCreateFarrowingUseCase{farrowing: farrowing}
		handler := newTestHandler(create, &fakeListFarrowingsUseCase{})
		invalid := fmt.Sprintf(
			`{"sow_id":"%s","farrow_date":"2026-04-20","extra":true}`,
			sowID,
		)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/farrowings", strings.NewReader(invalid))
		request = authenticatedRequest(request, actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		create := &fakeCreateFarrowingUseCase{farrowing: farrowing}
		handler := newTestHandler(create, &fakeListFarrowingsUseCase{})
		response := serve(handler, httptest.NewRequest(http.MethodPost, "/api/v1/farrowings", strings.NewReader(body)))

		if response.Code != http.StatusUnauthorized || create.called {
			t.Fatalf("status=%d called=%v", response.Code, create.called)
		}
	})

	t.Run("rejects a missing piglet count", func(t *testing.T) {
		create := &fakeCreateFarrowingUseCase{farrowing: farrowing}
		handler := newTestHandler(create, &fakeListFarrowingsUseCase{})

		for _, invalid := range []string{
			fmt.Sprintf(`{"sow_id":"%s","farrow_date":"2026-04-20","stillborn":1,"mummified":0}`, sowID),
			fmt.Sprintf(`{"sow_id":"%s","farrow_date":"2026-04-20","live_born":10,"mummified":0}`, sowID),
			fmt.Sprintf(`{"sow_id":"%s","farrow_date":"2026-04-20","live_born":10,"stillborn":1}`, sowID),
			fmt.Sprintf(`{"sow_id":"%s","farrow_date":"2026-04-20","live_born":null,"stillborn":1,"mummified":0}`, sowID),
		} {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/farrowings", strings.NewReader(invalid))
			request = authenticatedRequest(request, actorID)
			response := serve(handler, request)
			if response.Code != http.StatusBadRequest || create.called {
				t.Fatalf("status=%d called=%v body=%s", response.Code, create.called, invalid)
			}
		}
	})

	t.Run("returns bad request for invalid fields", func(t *testing.T) {
		create := &fakeCreateFarrowingUseCase{farrowing: farrowing}
		handler := newTestHandler(create, &fakeListFarrowingsUseCase{})

		for _, invalid := range []string{
			`{"sow_id":"not-a-uuid","farrow_date":"2026-04-20"}`,
			fmt.Sprintf(`{"sow_id":"%s","farrow_date":"not-a-date"}`, sowID),
			fmt.Sprintf(`{"sow_id":"%s","farrow_date":"2026-04-20","operators":[{"operator_id":"bad"}]}`, sowID),
			fmt.Sprintf(`{"sow_id":"%s","farrow_date":"2026-04-20","medications":[{"medication_id":"bad","dose":1,"applied_by":"%s"}]}`, sowID, operatorID),
		} {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/farrowings", strings.NewReader(invalid))
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
			{"sow not found", ports.ErrSowNotFound, http.StatusNotFound},
			{"service not found", ports.ErrServiceNotFound, http.StatusNotFound},
			{"operator not found", ports.ErrOperatorNotFound, http.StatusNotFound},
			{"medication not found", ports.ErrMedicationNotFound, http.StatusNotFound},
			{"sow not gestating", farrowingapplication.ErrSowNotGestating, http.StatusConflict},
			{"service not confirmed", farrowingapplication.ErrServiceNotConfirmed, http.StatusConflict},
			{"already exists", ports.ErrFarrowingAlreadyExists, http.StatusConflict},
			{"date before mount", farrowingdomain.ErrFarrowDateBeforeMount, http.StatusBadRequest},
			{"date in future", farrowingdomain.ErrFarrowDateInFuture, http.StatusBadRequest},
			{"duplicate operator", farrowingdomain.ErrDuplicateOperator, http.StatusBadRequest},
			{"invalid dose", farrowingdomain.ErrInvalidFarrowingMedicationDose, http.StatusBadRequest},
			{"live born too high", farrowingdomain.ErrLiveBornTooHigh, http.StatusBadRequest},
			{"stillborn too high", farrowingdomain.ErrStillbornTooHigh, http.StatusBadRequest},
			{"mummified too high", farrowingdomain.ErrMummifiedTooHigh, http.StatusBadRequest},
			{"live born required", farrowingdomain.ErrLiveBornRequired, http.StatusBadRequest},
			{"stillborn required", farrowingdomain.ErrStillbornRequired, http.StatusBadRequest},
			{"mummified required", farrowingdomain.ErrMummifiedRequired, http.StatusBadRequest},
			{"litter weight required", farrowingdomain.ErrLitterWeightRequired, http.StatusBadRequest},
			{"stillborn weight required", farrowingdomain.ErrStillbornWeightRequired, http.StatusBadRequest},
			{"litter weight too high", farrowingdomain.ErrLitterWeightTooHigh, http.StatusBadRequest},
			{"stillborn weight too high", farrowingdomain.ErrStillbornWeightTooHigh, http.StatusBadRequest},
			{"medication dose too high", farrowingdomain.ErrMedicationDoseTooHigh, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				create := &fakeCreateFarrowingUseCase{err: test.err}
				handler := newTestHandler(create, &fakeListFarrowingsUseCase{})
				request := httptest.NewRequest(http.MethodPost, "/api/v1/farrowings", strings.NewReader(body))
				request = authenticatedRequest(request, actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestFarrowingHandlerList(t *testing.T) {
	t.Run("returns a page of farrowings", func(t *testing.T) {
		list := &fakeListFarrowingsUseCase{farrowings: []*farrowingdomain.Farrowing{handlerFarrowing()}}
		handler := newTestHandler(&fakeCreateFarrowingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/farrowings", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		body := response.Body.String()
		for _, want := range []string{`"items"`, `"total_pages"`, "2026-04-20", "Corral 3", "operators", "medications", "current_piglets", `"sow_code"`} {
			if !strings.Contains(body, want) {
				t.Fatalf("body missing %q: %s", want, body)
			}
		}
		for _, audit := range []string{"created_at", "updated_at", "created_by", "updated_by"} {
			if strings.Contains(body, audit) {
				t.Fatalf("audit field %q must be omitted: %s", audit, body)
			}
		}
	})

	t.Run("returns an empty page when there are no farrowings", func(t *testing.T) {
		list := &fakeListFarrowingsUseCase{farrowings: []*farrowingdomain.Farrowing{}}
		handler := newTestHandler(&fakeCreateFarrowingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/farrowings", nil))

		body := strings.TrimSpace(response.Body.String())
		if response.Code != http.StatusOK || !strings.Contains(body, `"items":[]`) || !strings.Contains(body, `"total":0`) {
			t.Fatalf("status=%d body=%s", response.Code, body)
		}
	})

	t.Run("parses the sow_code filter", func(t *testing.T) {
		list := &fakeListFarrowingsUseCase{}
		handler := newTestHandler(&fakeCreateFarrowingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(
			http.MethodGet,
			"/api/v1/farrowings?sow_code=C-001",
			nil,
		))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		if list.filter.SowCode == nil || *list.filter.SowCode != "C-001" {
			t.Fatalf("unexpected sow code: %#v", list.filter.SowCode)
		}
	})

	t.Run("parses the from and to date range filters", func(t *testing.T) {
		list := &fakeListFarrowingsUseCase{}
		handler := newTestHandler(&fakeCreateFarrowingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(
			http.MethodGet,
			"/api/v1/farrowings?from=2026-04-01&to=2026-04-30",
			nil,
		))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		if list.filter.FromDate == nil || !list.filter.FromDate.Equal(time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected from date: %#v", list.filter.FromDate)
		}
		if list.filter.ToDate == nil || !list.filter.ToDate.Equal(time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected to date: %#v", list.filter.ToDate)
		}
	})

	t.Run("returns an empty filter when no params are given", func(t *testing.T) {
		list := &fakeListFarrowingsUseCase{}
		handler := newTestHandler(&fakeCreateFarrowingUseCase{}, list)
		serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/farrowings", nil))

		if list.filter.SowCode != nil || list.filter.FromDate != nil || list.filter.ToDate != nil {
			t.Fatalf("unexpected filter: %#v", list.filter)
		}
	})

	t.Run("parses the page query parameter", func(t *testing.T) {
		list := &fakeListFarrowingsUseCase{}
		handler := newTestHandler(&fakeCreateFarrowingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/farrowings?page=3", nil))

		if response.Code != http.StatusOK || list.page != 3 {
			t.Fatalf("status=%d page=%d", response.Code, list.page)
		}
	})

	t.Run("defaults the page to one", func(t *testing.T) {
		list := &fakeListFarrowingsUseCase{}
		handler := newTestHandler(&fakeCreateFarrowingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/farrowings", nil))

		if response.Code != http.StatusOK || list.page != 1 {
			t.Fatalf("status=%d page=%d", response.Code, list.page)
		}
	})

	t.Run("rejects invalid filter values", func(t *testing.T) {
		list := &fakeListFarrowingsUseCase{}
		handler := newTestHandler(&fakeCreateFarrowingUseCase{}, list)

		for _, query := range []string{
			"from=not-a-date",
			"to=not-a-date",
			"from=2026-04-30&to=2026-04-01",
			"page=0",
			"page=abc",
		} {
			response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/farrowings?"+query, nil))
			if response.Code != http.StatusBadRequest || list.called {
				t.Fatalf("status=%d called=%v query=%s", response.Code, list.called, query)
			}
		}
	})

	t.Run("maps internal errors", func(t *testing.T) {
		list := &fakeListFarrowingsUseCase{err: errors.New("unexpected")}
		handler := newTestHandler(&fakeCreateFarrowingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/farrowings", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want %d", response.Code, http.StatusInternalServerError)
		}
	})
}

func TestFarrowingHandlerUpdate(t *testing.T) {
	farrowingID := uuid.New()
	actorID := uuid.New()
	operatorID := uuid.New()
	medicationID := uuid.New()

	body := fmt.Sprintf(
		`{"farrow_date":"2026-05-01","live_born":12,"stillborn":0,"mummified":0,`+
			`"location":"Corral 5","clear_litter_weight":true,"is_manipulated":true,`+
			`"operators":[{"operator_id":"%s"}],`+
			`"medications":[{"medication_id":"%s","dose":2,"applied_by":"%s"}]}`,
		operatorID, medicationID, operatorID,
	)
	target := "/api/v1/farrowings/" + farrowingID.String()

	t.Run("applies the provided fields and returns 200 without a body", func(t *testing.T) {
		update := &fakeUpdateFarrowingUseCase{}
		handler := newUpdateHandler(update)
		request := authenticatedRequest(httptest.NewRequest(http.MethodPatch, target, strings.NewReader(body)), actorID)
		response := serve(handler, request)

		if response.Code != http.StatusOK || !update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
		if update.id != farrowingID || update.command.UpdatedBy != actorID {
			t.Fatalf("unexpected command: %#v", update.command)
		}
		if update.command.FarrowDate == nil || !update.command.FarrowDate.Equal(time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected date: %#v", update.command.FarrowDate)
		}
		if update.command.LiveBorn == nil || *update.command.LiveBorn != 12 {
			t.Fatalf("unexpected live born: %#v", update.command.LiveBorn)
		}
		if update.command.IsManipulated == nil || !*update.command.IsManipulated || !update.command.ClearLitterWeight {
			t.Fatalf("unexpected flags: %#v", update.command)
		}
		if !update.command.UpdateOperators || len(update.command.Operators) != 1 {
			t.Fatalf("unexpected operators: %#v", update.command)
		}
		if !update.command.UpdateMedications || len(update.command.Medications) != 1 {
			t.Fatalf("unexpected medications: %#v", update.command)
		}
		if response.Body.Len() != 0 {
			t.Fatalf("expected an empty body, got %s", response.Body.String())
		}
	})

	t.Run("accepts a partial body", func(t *testing.T) {
		update := &fakeUpdateFarrowingUseCase{}
		handler := newUpdateHandler(update)
		request := authenticatedRequest(
			httptest.NewRequest(http.MethodPatch, target, strings.NewReader(`{"location":"Corral 5"}`)),
			actorID,
		)
		response := serve(handler, request)

		if response.Code != http.StatusOK || !update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
		if update.command.Location == nil || *update.command.Location != "Corral 5" {
			t.Fatalf("unexpected location: %#v", update.command.Location)
		}
		if update.command.FarrowDate != nil || update.command.UpdateOperators || update.command.UpdateMedications {
			t.Fatalf("unexpected command: %#v", update.command)
		}
	})

	t.Run("rejects an invalid farrowing id", func(t *testing.T) {
		update := &fakeUpdateFarrowingUseCase{}
		handler := newUpdateHandler(update)
		request := authenticatedRequest(
			httptest.NewRequest(http.MethodPatch, "/api/v1/farrowings/not-a-uuid", strings.NewReader(`{"location":"X"}`)),
			actorID,
		)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("rejects invalid fields", func(t *testing.T) {
		update := &fakeUpdateFarrowingUseCase{}
		handler := newUpdateHandler(update)

		for _, invalid := range []string{
			`{"farrow_date":"not-a-date"}`,
			`{"operators":[{"operator_id":"bad"}]}`,
			`{"medications":[{"medication_id":"bad","dose":1,"applied_by":"bad"}]}`,
		} {
			request := authenticatedRequest(httptest.NewRequest(http.MethodPatch, target, strings.NewReader(invalid)), actorID)
			response := serve(handler, request)
			if response.Code != http.StatusBadRequest || update.called {
				t.Fatalf("status=%d called=%v body=%s", response.Code, update.called, invalid)
			}
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		update := &fakeUpdateFarrowingUseCase{}
		handler := newUpdateHandler(update)
		response := serve(handler, httptest.NewRequest(http.MethodPatch, target, strings.NewReader(body)))

		if response.Code != http.StatusUnauthorized || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("blocks the update for non-admins", func(t *testing.T) {
		update := &fakeUpdateFarrowingUseCase{}
		handler := newAdminGuardUpdateHandler(update)
		request := authenticatedRequest(httptest.NewRequest(http.MethodPatch, target, strings.NewReader(body)), actorID)
		response := serve(handler, request)

		if response.Code != http.StatusForbidden || update.called {
			t.Fatalf("status=%d called=%v", response.Code, update.called)
		}
	})

	t.Run("maps errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"not found", ports.ErrFarrowingNotFound, http.StatusNotFound},
			{"invalid update", farrowingdomain.ErrInvalidUpdate, http.StatusBadRequest},
			{"live born locked", farrowingdomain.ErrLiveBornLocked, http.StatusBadRequest},
			{"date after related", farrowingdomain.ErrFarrowDateAfterRelatedRecords, http.StatusBadRequest},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				update := &fakeUpdateFarrowingUseCase{err: test.err}
				handler := newUpdateHandler(update)
				request := authenticatedRequest(httptest.NewRequest(http.MethodPatch, target, strings.NewReader(body)), actorID)
				response := serve(handler, request)
				if response.Code != test.status {
					t.Fatalf("status=%d, want %d", response.Code, test.status)
				}
			})
		}
	})
}

func TestFarrowingHandlerDelete(t *testing.T) {
	farrowingID := uuid.New()
	actorID := uuid.New()
	target := "/api/v1/farrowings/" + farrowingID.String()

	t.Run("deletes the farrowing and returns 204 without a body", func(t *testing.T) {
		remove := &fakeDeleteFarrowingUseCase{}
		handler := newDeleteHandler(remove)
		request := authenticatedRequest(httptest.NewRequest(http.MethodDelete, target, nil), actorID)
		response := serve(handler, request)

		if response.Code != http.StatusNoContent || !remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
		if remove.command.ID != farrowingID || remove.command.DeletedBy != actorID {
			t.Fatalf("unexpected command: %#v", remove.command)
		}
		if response.Body.Len() != 0 {
			t.Fatalf("expected an empty body, got %s", response.Body.String())
		}
	})

	t.Run("rejects an invalid id", func(t *testing.T) {
		remove := &fakeDeleteFarrowingUseCase{}
		handler := newDeleteHandler(remove)
		request := authenticatedRequest(httptest.NewRequest(http.MethodDelete, "/api/v1/farrowings/not-a-uuid", nil), actorID)
		response := serve(handler, request)

		if response.Code != http.StatusBadRequest || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("returns unauthorized when actor is missing", func(t *testing.T) {
		remove := &fakeDeleteFarrowingUseCase{}
		handler := newDeleteHandler(remove)
		response := serve(handler, httptest.NewRequest(http.MethodDelete, target, nil))

		if response.Code != http.StatusUnauthorized || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("blocks the delete for non-admins", func(t *testing.T) {
		remove := &fakeDeleteFarrowingUseCase{}
		handler := newAdminGuardDeleteHandler(remove)
		request := authenticatedRequest(httptest.NewRequest(http.MethodDelete, target, nil), actorID)
		response := serve(handler, request)

		if response.Code != http.StatusForbidden || remove.called {
			t.Fatalf("status=%d called=%v", response.Code, remove.called)
		}
	})

	t.Run("maps errors", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			err    error
			status int
		}{
			{"not found", ports.ErrFarrowingNotFound, http.StatusNotFound},
			{"related records", ports.ErrFarrowingHasRelatedRecords, http.StatusConflict},
			{"internal", errors.New("unexpected"), http.StatusInternalServerError},
		} {
			t.Run(test.name, func(t *testing.T) {
				remove := &fakeDeleteFarrowingUseCase{err: test.err}
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
