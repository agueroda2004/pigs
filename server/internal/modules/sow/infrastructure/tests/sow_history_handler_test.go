package tests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	abortiondomain "server/internal/modules/abortion/domain"
	farrowingdomain "server/internal/modules/farrowing/domain"
	servicedomain "server/internal/modules/service/domain"
	sowapplication "server/internal/modules/sow/application"
	sowdomain "server/internal/modules/sow/domain"
	sowinfra "server/internal/modules/sow/infrastructure"
	"server/internal/modules/sow/ports"
)

func serveHistory(handler *sowinfra.SowHistoryHandler, request *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func newHistoryHandler(
	sowUseCase *fakeSowHistoryBySowUseCase,
	serviceUseCase *fakeServicesHistoryBySowUseCase,
	abortionUseCase *fakeAbortionHistoryBySowUseCase,
	farrowingUseCase *fakeFarrowingHistoryBySowUseCase,
) *sowinfra.SowHistoryHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return sowinfra.NewSowHistoryHandler(sowUseCase, serviceUseCase, abortionUseCase, farrowingUseCase, passThrough)
}

func historySow() *sowdomain.SowHistoryBySow {
	birth := "2025-01-01"
	return &sowdomain.SowHistoryBySow{
		Code:      "C-001",
		EntryDate: "2026-01-10",
		BirthDate: &birth,
		State:     sowdomain.State("Viva"),
		Parity:    2,
		BreedName: "Large White",
		Age:       "1 año",
	}
}

func TestSowHistoryHandlerGet(t *testing.T) {
	t.Run("returns the unified event timeline", func(t *testing.T) {
		start := "22:00"
		end := "02:30"
		duration := "04:30"
		serviceUseCase := &fakeServicesHistoryBySowUseCase{services: []*servicedomain.ServiceHistoryBySow{
			{
				SowCode: "C-001",
				State:   servicedomain.StateConfirmed,
				Mounts: []servicedomain.ServiceHistoryMountBySow{
					{BoarCode: "B-001", OperatorName: "Op 1", MountDate: "2026-02-01", Type: servicedomain.MountTypeNatural},
				},
			},
		}}
		abortionUseCase := &fakeAbortionHistoryBySowUseCase{abortions: []abortiondomain.AbortionHistoryBySow{
			{SowCode: "C-001", AbortionDate: "2026-03-01", Cause: "Infeccioso"},
		}}
		farrowingUseCase := &fakeFarrowingHistoryBySowUseCase{farrowings: []*farrowingdomain.FarrowingHistoryBySow{
			{
				FarrowDate:    "2026-04-01",
				StartTime:     &start,
				EndTime:       &end,
				Duration:      &duration,
				LiveBorn:      10,
				Stillborn:     1,
				Mummified:     1,
				IsManipulated: true,
				IsNurse:       true,
				Operators: []farrowingdomain.FarrowingHistoryOperator{
					{OperatorName: "Op 1"},
				},
				Medications: []farrowingdomain.FarrowingHistoryMedication{
					{MedicationName: "Oxitocina", Dose: 2, AppliedByName: "Op 2"},
				},
			},
		}}
		sowUseCase := &fakeSowHistoryBySowUseCase{history: historySow()}
		handler := newHistoryHandler(sowUseCase, serviceUseCase, abortionUseCase, farrowingUseCase)

		response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/c-001/history", nil))

		if response.Code != http.StatusOK {
			t.Fatalf("status=%d", response.Code)
		}
		if !sowUseCase.called || !serviceUseCase.called || !abortionUseCase.called || !farrowingUseCase.called {
			t.Fatalf("not every use case was called")
		}
		if farrowingUseCase.code != "c-001" {
			t.Fatalf("farrowing code = %q, want c-001", farrowingUseCase.code)
		}
		body := response.Body.String()
		for _, want := range []string{
			`"sow"`, `"events"`, `"type":"sow_entry"`, `"type":"service"`, `"type":"abortion"`, `"type":"farrowing"`,
			`"code":"C-001"`, `"entry_date":"2026-01-10"`, `"birth_date":"2025-01-01"`,
			`"breed_name":"Large White"`, `"age":"1 año"`,
			`"boar_code":"B-001"`, `"mount_date":"2026-02-01"`,
			`"abortion_date":"2026-03-01"`, `"cause":"Infeccioso"`,
			`"farrow_date":"2026-04-01"`, `"start_time":"22:00"`, `"end_time":"02:30"`, `"duration":"04:30"`,
			`"live_born":10`, `"stillborn":1`, `"mummified":1`,
			`"is_manipulated":true`, `"is_nurse":true`,
			`"operator_name":"Op 1"`, `"medication_name":"Oxitocina"`, `"dose":2`, `"applied_by_name":"Op 2"`,
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("body missing %q: %s", want, body)
			}
		}
		farrowingIndex := strings.Index(body, `"type":"farrowing"`)
		abortionIndex := strings.Index(body, `"type":"abortion"`)
		serviceIndex := strings.Index(body, `"type":"service"`)
		entryIndex := strings.Index(body, `"type":"sow_entry"`)
		if farrowingIndex > abortionIndex || abortionIndex > serviceIndex || serviceIndex > entryIndex {
			t.Fatalf("events are not newest-first: %s", body)
		}
		if strings.Contains(body, `"services"`) || strings.Contains(body, `"abortions"`) {
			t.Fatalf("legacy arrays must be absent: %s", body)
		}
	})

	t.Run("returns only the sow entry when there is no reproductive history", func(t *testing.T) {
		sowUseCase := &fakeSowHistoryBySowUseCase{history: historySow()}
		handler := newHistoryHandler(
			sowUseCase,
			&fakeServicesHistoryBySowUseCase{services: nil},
			&fakeAbortionHistoryBySowUseCase{},
			&fakeFarrowingHistoryBySowUseCase{},
		)

		response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/c-001/history", nil))

		if response.Code != http.StatusOK {
			t.Fatalf("status=%d", response.Code)
		}
		body := response.Body.String()
		if !strings.Contains(body, `"events":[{"type":"sow_entry"`) {
			t.Fatalf("expected only the sow entry event: %s", body)
		}
		if strings.Contains(body, `"type":"service"`) || strings.Contains(body, `"type":"abortion"`) || strings.Contains(body, `"type":"farrowing"`) {
			t.Fatalf("unexpected reproductive events: %s", body)
		}
	})

	t.Run("returns not found for missing or inactive sow", func(t *testing.T) {
		for _, testErr := range []error{ports.ErrSowNotFound, sowapplication.ErrSowInactive} {
			serviceUseCase := &fakeServicesHistoryBySowUseCase{}
			abortionUseCase := &fakeAbortionHistoryBySowUseCase{}
			farrowingUseCase := &fakeFarrowingHistoryBySowUseCase{}
			sowUseCase := &fakeSowHistoryBySowUseCase{err: testErr}
			handler := newHistoryHandler(sowUseCase, serviceUseCase, abortionUseCase, farrowingUseCase)

			response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/C-001/history", nil))

			if response.Code != http.StatusNotFound {
				t.Fatalf("status=%d", response.Code)
			}
			if serviceUseCase.called || abortionUseCase.called || farrowingUseCase.called {
				t.Fatal("history use cases must not be called when the sow fails")
			}
		}
	})

	t.Run("returns internal error for dependency failure", func(t *testing.T) {
		sowFailure := &fakeSowHistoryBySowUseCase{err: errors.New("db down")}
		handler := newHistoryHandler(sowFailure, &fakeServicesHistoryBySowUseCase{}, &fakeAbortionHistoryBySowUseCase{}, &fakeFarrowingHistoryBySowUseCase{})
		if response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/C-001/history", nil)); response.Code != http.StatusInternalServerError {
			t.Fatalf("sow failure status=%d", response.Code)
		}

		serviceFailure := &fakeServicesHistoryBySowUseCase{err: errors.New("db down")}
		handler = newHistoryHandler(&fakeSowHistoryBySowUseCase{history: historySow()}, serviceFailure, &fakeAbortionHistoryBySowUseCase{}, &fakeFarrowingHistoryBySowUseCase{})
		if response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/C-001/history", nil)); response.Code != http.StatusInternalServerError {
			t.Fatalf("service failure status=%d", response.Code)
		}

		abortionFailure := &fakeAbortionHistoryBySowUseCase{err: errors.New("db down")}
		handler = newHistoryHandler(&fakeSowHistoryBySowUseCase{history: historySow()}, &fakeServicesHistoryBySowUseCase{}, abortionFailure, &fakeFarrowingHistoryBySowUseCase{})
		if response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/C-001/history", nil)); response.Code != http.StatusInternalServerError {
			t.Fatalf("abortion failure status=%d", response.Code)
		}

		farrowingFailure := &fakeFarrowingHistoryBySowUseCase{err: errors.New("db down")}
		handler = newHistoryHandler(&fakeSowHistoryBySowUseCase{history: historySow()}, &fakeServicesHistoryBySowUseCase{}, &fakeAbortionHistoryBySowUseCase{}, farrowingFailure)
		if response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/C-001/history", nil)); response.Code != http.StatusInternalServerError {
			t.Fatalf("farrowing failure status=%d", response.Code)
		}
	})
}
