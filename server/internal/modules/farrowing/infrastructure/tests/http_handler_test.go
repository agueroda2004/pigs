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
		if !create.command.FarrowDate.Equal(time.Date(2026, time.April, 20, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected date: %v", create.command.FarrowDate)
		}
		if len(create.command.Operators) != 1 || create.command.Operators[0].OperatorID != operatorID {
			t.Fatalf("unexpected operators: %#v", create.command.Operators)
		}
		if len(create.command.Medications) != 1 || create.command.Medications[0].MedicationID != medicationID {
			t.Fatalf("unexpected medications: %#v", create.command.Medications)
		}
		if response.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type: %q", response.Header().Get("Content-Type"))
		}
		if !strings.Contains(response.Body.String(), "Corral 3") {
			t.Fatalf("unexpected body: %s", response.Body.String())
		}
	})

	t.Run("accepts a decimal medication dose", func(t *testing.T) {
		create := &fakeCreateFarrowingUseCase{farrowing: farrowing}
		handler := newTestHandler(create, &fakeListFarrowingsUseCase{})
		decimalBody := fmt.Sprintf(
			`{"sow_id":"%s","farrow_date":"2026-04-20",`+
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
	t.Run("returns every farrowing", func(t *testing.T) {
		list := &fakeListFarrowingsUseCase{farrowings: []*farrowingdomain.Farrowing{handlerFarrowing()}}
		handler := newTestHandler(&fakeCreateFarrowingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/farrowings", nil))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		body := response.Body.String()
		for _, want := range []string{"2026-04-20", "Corral 3", "operators", "medications"} {
			if !strings.Contains(body, want) {
				t.Fatalf("body missing %q: %s", want, body)
			}
		}
	})

	t.Run("returns an empty array when there are no farrowings", func(t *testing.T) {
		list := &fakeListFarrowingsUseCase{farrowings: []*farrowingdomain.Farrowing{}}
		handler := newTestHandler(&fakeCreateFarrowingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(http.MethodGet, "/api/v1/farrowings", nil))

		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("parses the sow_id and service_id filters", func(t *testing.T) {
		sowID := uuid.New()
		serviceID := uuid.New()
		list := &fakeListFarrowingsUseCase{}
		handler := newTestHandler(&fakeCreateFarrowingUseCase{}, list)
		response := serve(handler, httptest.NewRequest(
			http.MethodGet,
			"/api/v1/farrowings?sow_id="+sowID.String()+"&service_id="+serviceID.String(),
			nil,
		))

		if response.Code != http.StatusOK || !list.called {
			t.Fatalf("status=%d called=%v", response.Code, list.called)
		}
		if list.filter.SowID == nil || *list.filter.SowID != sowID {
			t.Fatalf("unexpected sow id: %#v", list.filter.SowID)
		}
		if list.filter.ServiceID == nil || *list.filter.ServiceID != serviceID {
			t.Fatalf("unexpected service id: %#v", list.filter.ServiceID)
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

		if list.filter.SowID != nil || list.filter.ServiceID != nil ||
			list.filter.FromDate != nil || list.filter.ToDate != nil {
			t.Fatalf("unexpected filter: %#v", list.filter)
		}
	})

	t.Run("rejects invalid filter values", func(t *testing.T) {
		list := &fakeListFarrowingsUseCase{}
		handler := newTestHandler(&fakeCreateFarrowingUseCase{}, list)

		for _, query := range []string{
			"sow_id=not-a-uuid",
			"service_id=not-a-uuid",
			"from=not-a-date",
			"to=not-a-date",
			"from=2026-04-30&to=2026-04-01",
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
