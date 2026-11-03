package tests

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	abortiondomain "server/internal/modules/abortion/domain"
	farrowingdomain "server/internal/modules/farrowing/domain"
	partialweagingdomain "server/internal/modules/partialweaging/domain"
	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
	servicedomain "server/internal/modules/service/domain"
	sowapplication "server/internal/modules/sow/application"
	sowdomain "server/internal/modules/sow/domain"
	sowinfra "server/internal/modules/sow/infrastructure"
	"server/internal/modules/sow/ports"
	weagingdomain "server/internal/modules/weaging/domain"
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
	pigletDeathUseCase *fakePigletDeathHistoryBySowUseCase,
	pigletFosteringUseCase *fakePigletFosteringHistoryBySowUseCase,
	partialWeagingUseCase *fakePartialWeagingHistoryBySowUseCase,
	weagingUseCase *fakeWeagingHistoryBySowUseCase,
) *sowinfra.SowHistoryHandler {
	passThrough := func(next http.Handler) http.Handler { return next }
	return sowinfra.NewSowHistoryHandler(
		sowUseCase, serviceUseCase, abortionUseCase, farrowingUseCase,
		pigletDeathUseCase, pigletFosteringUseCase, partialWeagingUseCase, weagingUseCase, passThrough,
	)
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
		note := "Muerte durante la noche"
		weight := 2.5
		fosteringNote := "Traslado a la cerda 106"
		sowID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
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
		pigletDeathUseCase := &fakePigletDeathHistoryBySowUseCase{deaths: []*pigletdeathdomain.PigletDeathHistoryBySow{
			{
				OperatorName: "Op 3",
				DeathDate:    time.Date(2026, time.April, 5, 0, 0, 0, 0, time.UTC),
				Quantity:     3,
				Weight:       &weight,
				Cause:        pigletdeathdomain.CauseCrushed,
				Turn:         pigletdeathdomain.TurnMorning,
				Note:         &note,
			},
		}}
		pigletFosteringUseCase := &fakePigletFosteringHistoryBySowUseCase{fosterings: []*pigletfosteringdomain.PigletFosteringHistoryBySow{
			{
				SowID:        sowID,
				Role:         pigletfosteringdomain.PigletFosteringRoleDonor,
				MovementDate: time.Date(2026, time.April, 10, 0, 0, 0, 0, time.UTC),
				Quantity:     -2,
				Note:         &fosteringNote,
			},
		}}
		partialWeagingNote := "Destete parcial de la camada"
		partialWeagingWeight := 42.5
		partialWeagingUseCase := &fakePartialWeagingHistoryBySowUseCase{weagings: []*partialweagingdomain.PartialWeagingHistoryBySow{
			{
				WeagingDate: time.Date(2026, time.April, 15, 0, 0, 0, 0, time.UTC),
				Quantity:    2,
				TotalWeight: &partialWeagingWeight,
				Type:        partialweagingdomain.TypeNodriza,
				Note:        &partialWeagingNote,
			},
		}}
		weagingNote := "Destete completo"
		weagingWeight := 55.5
		weagingUseCase := &fakeWeagingHistoryBySowUseCase{weagings: []*weagingdomain.WeagingHistoryBySow{
			{
				WeagingDate: time.Date(2026, time.April, 20, 0, 0, 0, 0, time.UTC),
				Quantity:    5,
				TotalWeight: &weagingWeight,
				Note:        &weagingNote,
			},
		}}
		sowUseCase := &fakeSowHistoryBySowUseCase{history: historySow()}
		handler := newHistoryHandler(
			sowUseCase, serviceUseCase, abortionUseCase, farrowingUseCase,
			pigletDeathUseCase, pigletFosteringUseCase, partialWeagingUseCase, weagingUseCase,
		)

		response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/c-001/history", nil))

		if response.Code != http.StatusOK {
			t.Fatalf("status=%d", response.Code)
		}
		if !sowUseCase.called || !serviceUseCase.called || !abortionUseCase.called ||
			!farrowingUseCase.called || !pigletDeathUseCase.called || !pigletFosteringUseCase.called ||
			!partialWeagingUseCase.called || !weagingUseCase.called {
			t.Fatalf("not every use case was called")
		}
		if farrowingUseCase.code != "c-001" {
			t.Fatalf("farrowing code = %q, want c-001", farrowingUseCase.code)
		}
		if pigletDeathUseCase.code != "c-001" {
			t.Fatalf("piglet death code = %q, want c-001", pigletDeathUseCase.code)
		}
		if pigletFosteringUseCase.code != "c-001" {
			t.Fatalf("piglet fostering code = %q, want c-001", pigletFosteringUseCase.code)
		}
		if partialWeagingUseCase.code != "c-001" {
			t.Fatalf("partial weaging code = %q, want c-001", partialWeagingUseCase.code)
		}
		if weagingUseCase.code != "c-001" {
			t.Fatalf("weaging code = %q, want c-001", weagingUseCase.code)
		}
		body := response.Body.String()
		for _, want := range []string{
			`"sow"`, `"events"`, `"type":"sow_entry"`, `"type":"service"`, `"type":"abortion"`, `"type":"farrowing"`,
			`"type":"piglet_death"`, `"type":"piglet_fostering"`, `"type":"partial_weaging"`, `"type":"weaging"`,
			`"code":"C-001"`, `"entry_date":"2026-01-10"`, `"birth_date":"2025-01-01"`,
			`"breed_name":"Large White"`, `"age":"1 año"`,
			`"boar_code":"B-001"`, `"mount_date":"2026-02-01"`,
			`"abortion_date":"2026-03-01"`, `"cause":"Infeccioso"`,
			`"farrow_date":"2026-04-01"`, `"start_time":"22:00"`, `"end_time":"02:30"`, `"duration":"04:30"`,
			`"live_born":10`, `"stillborn":1`, `"mummified":1`,
			`"is_manipulated":true`, `"is_nurse":true`,
			`"operator_name":"Op 1"`, `"medication_name":"Oxitocina"`, `"dose":2`, `"applied_by_name":"Op 2"`,
			`"operator_name":"Op 3"`, `"death_date":"2026-04-05"`, `"quantity":3`, `"weight":2.5`,
			`"cause":"Aplastado"`, `"turn":"Mañana"`, `"note":"Muerte durante la noche"`,
			`"sow_id":"11111111-1111-1111-1111-111111111111"`, `"role":"donor"`,
			`"movement_date":"2026-04-10"`, `"quantity":-2`, `"note":"Traslado a la cerda 106"`,
			`"weaging_date":"2026-04-15"`, `"total_weight":42.5`, `"type":"Nodriza"`,
			`"note":"Destete parcial de la camada"`,
			`"weaging_date":"2026-04-20"`, `"quantity":5`, `"total_weight":55.5`,
			`"note":"Destete completo"`,
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("body missing %q: %s", want, body)
			}
		}
		weagingIndex := strings.Index(body, `"type":"weaging"`)
		partialWeagingIndex := strings.Index(body, `"type":"partial_weaging"`)
		fosteringIndex := strings.Index(body, `"type":"piglet_fostering"`)
		pigletDeathIndex := strings.Index(body, `"type":"piglet_death"`)
		farrowingIndex := strings.Index(body, `"type":"farrowing"`)
		abortionIndex := strings.Index(body, `"type":"abortion"`)
		serviceIndex := strings.Index(body, `"type":"service"`)
		entryIndex := strings.Index(body, `"type":"sow_entry"`)
		if weagingIndex > partialWeagingIndex || partialWeagingIndex > fosteringIndex ||
			fosteringIndex > pigletDeathIndex || pigletDeathIndex > farrowingIndex ||
			farrowingIndex > abortionIndex || abortionIndex > serviceIndex || serviceIndex > entryIndex {
			t.Fatalf("events are not newest-first: %s", body)
		}
		if strings.Contains(body, `"services"`) || strings.Contains(body, `"abortions"`) {
			t.Fatalf("legacy arrays must be absent: %s", body)
		}
	})

	t.Run("exposes exactly the seven piglet death fields with nullable weight and note", func(t *testing.T) {
		sowUseCase := &fakeSowHistoryBySowUseCase{history: historySow()}
		pigletDeathUseCase := &fakePigletDeathHistoryBySowUseCase{deaths: []*pigletdeathdomain.PigletDeathHistoryBySow{
			{
				OperatorName: "Op 3",
				DeathDate:    time.Date(2026, time.April, 5, 0, 0, 0, 0, time.UTC),
				Quantity:     3,
				Weight:       nil,
				Cause:        pigletdeathdomain.CauseCrushed,
				Turn:         pigletdeathdomain.TurnMorning,
				Note:         nil,
			},
		}}
		handler := newHistoryHandler(
			sowUseCase,
			&fakeServicesHistoryBySowUseCase{},
			&fakeAbortionHistoryBySowUseCase{},
			&fakeFarrowingHistoryBySowUseCase{},
			pigletDeathUseCase,
			&fakePigletFosteringHistoryBySowUseCase{},
			&fakePartialWeagingHistoryBySowUseCase{},
			&fakeWeagingHistoryBySowUseCase{},
		)

		response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/C-001/history", nil))

		if response.Code != http.StatusOK {
			t.Fatalf("status=%d", response.Code)
		}

		var payload struct {
			Events []struct {
				Type string          `json:"type"`
				Date string          `json:"date"`
				Data json.RawMessage `json:"data"`
			} `json:"events"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}

		var data map[string]any
		for _, event := range payload.Events {
			if event.Type == "piglet_death" {
				if event.Date != "2026-04-05" {
					t.Fatalf("event date = %q, want 2026-04-05", event.Date)
				}
				if err := json.Unmarshal(event.Data, &data); err != nil {
					t.Fatalf("invalid piglet death data: %v", err)
				}
				break
			}
		}
		if data == nil {
			t.Fatalf("missing piglet death event: %s", response.Body.String())
		}
		if len(data) != 7 {
			t.Fatalf("piglet death fields = %v, want exactly seven", data)
		}
		for _, key := range []string{"operator_name", "death_date", "quantity", "weight", "cause", "turn", "note"} {
			if _, ok := data[key]; !ok {
				t.Fatalf("piglet death missing %q: %v", key, data)
			}
		}
		if data["note"] != nil {
			t.Fatalf("note = %v, want null", data["note"])
		}
		if data["weight"] != nil {
			t.Fatalf("weight = %v, want null", data["weight"])
		}
		if !strings.Contains(response.Body.String(), `"weight":null`) {
			t.Fatalf("body must contain a null weight: %s", response.Body.String())
		}
	})

	t.Run("exposes exactly the five fostering fields with the requested sow perspective", func(t *testing.T) {
		sowID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		sowUseCase := &fakeSowHistoryBySowUseCase{history: historySow()}
		pigletFosteringUseCase := &fakePigletFosteringHistoryBySowUseCase{fosterings: []*pigletfosteringdomain.PigletFosteringHistoryBySow{
			{
				SowID:        sowID,
				Role:         pigletfosteringdomain.PigletFosteringRoleReceiver,
				MovementDate: time.Date(2026, time.April, 10, 0, 0, 0, 0, time.UTC),
				Quantity:     3,
				Note:         nil,
			},
		}}
		handler := newHistoryHandler(
			sowUseCase,
			&fakeServicesHistoryBySowUseCase{},
			&fakeAbortionHistoryBySowUseCase{},
			&fakeFarrowingHistoryBySowUseCase{},
			&fakePigletDeathHistoryBySowUseCase{},
			pigletFosteringUseCase,
			&fakePartialWeagingHistoryBySowUseCase{},
			&fakeWeagingHistoryBySowUseCase{},
		)

		response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/C-001/history", nil))

		if response.Code != http.StatusOK {
			t.Fatalf("status=%d", response.Code)
		}

		var payload struct {
			Events []struct {
				Type string          `json:"type"`
				Date string          `json:"date"`
				Data json.RawMessage `json:"data"`
			} `json:"events"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}

		var data map[string]any
		for _, event := range payload.Events {
			if event.Type == "piglet_fostering" {
				if event.Date != "2026-04-10" {
					t.Fatalf("event date = %q, want 2026-04-10", event.Date)
				}
				if err := json.Unmarshal(event.Data, &data); err != nil {
					t.Fatalf("invalid fostering data: %v", err)
				}
				break
			}
		}
		if data == nil {
			t.Fatalf("missing fostering event: %s", response.Body.String())
		}
		if len(data) != 5 {
			t.Fatalf("fostering fields = %v, want exactly five", data)
		}
		for _, key := range []string{"sow_id", "role", "movement_date", "quantity", "note"} {
			if _, ok := data[key]; !ok {
				t.Fatalf("fostering missing %q: %v", key, data)
			}
		}
		if data["sow_id"] != "22222222-2222-2222-2222-222222222222" {
			t.Fatalf("sow_id = %v, want the requested sow", data["sow_id"])
		}
		if data["role"] != "receiver" {
			t.Fatalf("role = %v, want receiver", data["role"])
		}
		if data["movement_date"] != "2026-04-10" {
			t.Fatalf("movement_date = %v, want 2026-04-10", data["movement_date"])
		}
		if data["quantity"] != float64(3) {
			t.Fatalf("quantity = %v, want 3", data["quantity"])
		}
		if data["note"] != nil {
			t.Fatalf("note = %v, want null", data["note"])
		}
	})

	t.Run("exposes exactly the five partial weaging fields with nullable values", func(t *testing.T) {
		sowUseCase := &fakeSowHistoryBySowUseCase{history: historySow()}
		partialWeagingUseCase := &fakePartialWeagingHistoryBySowUseCase{weagings: []*partialweagingdomain.PartialWeagingHistoryBySow{
			{
				WeagingDate: time.Date(2026, time.April, 15, 0, 0, 0, 0, time.UTC),
				Quantity:    4,
				TotalWeight: nil,
				Type:        partialweagingdomain.TypeBajaViabilidad,
				Note:        nil,
			},
		}}
		handler := newHistoryHandler(
			sowUseCase,
			&fakeServicesHistoryBySowUseCase{},
			&fakeAbortionHistoryBySowUseCase{},
			&fakeFarrowingHistoryBySowUseCase{},
			&fakePigletDeathHistoryBySowUseCase{},
			&fakePigletFosteringHistoryBySowUseCase{},
			partialWeagingUseCase,
			&fakeWeagingHistoryBySowUseCase{},
		)

		response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/C-001/history", nil))

		if response.Code != http.StatusOK {
			t.Fatalf("status=%d", response.Code)
		}

		var payload struct {
			Events []struct {
				Type string          `json:"type"`
				Date string          `json:"date"`
				Data json.RawMessage `json:"data"`
			} `json:"events"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}

		var data map[string]any
		for _, event := range payload.Events {
			if event.Type == "partial_weaging" {
				if event.Date != "2026-04-15" {
					t.Fatalf("event date = %q, want 2026-04-15", event.Date)
				}
				if err := json.Unmarshal(event.Data, &data); err != nil {
					t.Fatalf("invalid partial weaging data: %v", err)
				}
				break
			}
		}
		if data == nil {
			t.Fatalf("missing partial weaging event: %s", response.Body.String())
		}
		if len(data) != 5 {
			t.Fatalf("partial weaging fields = %v, want exactly five", data)
		}
		for _, key := range []string{"weaging_date", "quantity", "total_weight", "type", "note"} {
			if _, ok := data[key]; !ok {
				t.Fatalf("partial weaging missing %q: %v", key, data)
			}
		}
		if data["weaging_date"] != "2026-04-15" {
			t.Fatalf("weaging_date = %v, want 2026-04-15", data["weaging_date"])
		}
		if data["quantity"] != float64(4) {
			t.Fatalf("quantity = %v, want 4", data["quantity"])
		}
		if data["type"] != "Baja_Viabilidad" {
			t.Fatalf("type = %v, want Baja_Viabilidad", data["type"])
		}
		if data["total_weight"] != nil {
			t.Fatalf("total_weight = %v, want null", data["total_weight"])
		}
		if data["note"] != nil {
			t.Fatalf("note = %v, want null", data["note"])
		}
		if !strings.Contains(response.Body.String(), `"total_weight":null`) {
			t.Fatalf("body must contain a null total weight: %s", response.Body.String())
		}
	})

	t.Run("exposes exactly the four weaging fields with nullable values", func(t *testing.T) {
		weight := 55.5
		sowUseCase := &fakeSowHistoryBySowUseCase{history: historySow()}
		weagingUseCase := &fakeWeagingHistoryBySowUseCase{weagings: []*weagingdomain.WeagingHistoryBySow{
			{
				WeagingDate: time.Date(2026, time.April, 20, 0, 0, 0, 0, time.UTC),
				Quantity:    5,
				TotalWeight: &weight,
				Note:        nil,
			},
		}}
		handler := newHistoryHandler(
			sowUseCase,
			&fakeServicesHistoryBySowUseCase{},
			&fakeAbortionHistoryBySowUseCase{},
			&fakeFarrowingHistoryBySowUseCase{},
			&fakePigletDeathHistoryBySowUseCase{},
			&fakePigletFosteringHistoryBySowUseCase{},
			&fakePartialWeagingHistoryBySowUseCase{},
			weagingUseCase,
		)

		response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/C-001/history", nil))

		if response.Code != http.StatusOK {
			t.Fatalf("status=%d", response.Code)
		}

		var payload struct {
			Events []struct {
				Type string          `json:"type"`
				Date string          `json:"date"`
				Data json.RawMessage `json:"data"`
			} `json:"events"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}

		var data map[string]any
		for _, event := range payload.Events {
			if event.Type == "weaging" {
				if event.Date != "2026-04-20" {
					t.Fatalf("event date = %q, want 2026-04-20", event.Date)
				}
				if err := json.Unmarshal(event.Data, &data); err != nil {
					t.Fatalf("invalid weaging data: %v", err)
				}
				break
			}
		}
		if data == nil {
			t.Fatalf("missing weaging event: %s", response.Body.String())
		}
		if len(data) != 4 {
			t.Fatalf("weaging fields = %v, want exactly four", data)
		}
		for _, key := range []string{"weaging_date", "quantity", "total_weight", "note"} {
			if _, ok := data[key]; !ok {
				t.Fatalf("weaging missing %q: %v", key, data)
			}
		}
		if data["weaging_date"] != "2026-04-20" {
			t.Fatalf("weaging_date = %v, want 2026-04-20", data["weaging_date"])
		}
		if data["quantity"] != float64(5) {
			t.Fatalf("quantity = %v, want 5", data["quantity"])
		}
		if data["total_weight"] != 55.5 {
			t.Fatalf("total_weight = %v, want 55.5", data["total_weight"])
		}
		if data["note"] != nil {
			t.Fatalf("note = %v, want null", data["note"])
		}
		if !strings.Contains(response.Body.String(), `"total_weight":55.5`) {
			t.Fatalf("body must contain the total weight: %s", response.Body.String())
		}
	})

	t.Run("returns only the sow entry when there is no reproductive history", func(t *testing.T) {
		sowUseCase := &fakeSowHistoryBySowUseCase{history: historySow()}
		handler := newHistoryHandler(
			sowUseCase,
			&fakeServicesHistoryBySowUseCase{services: nil},
			&fakeAbortionHistoryBySowUseCase{},
			&fakeFarrowingHistoryBySowUseCase{},
			&fakePigletDeathHistoryBySowUseCase{},
			&fakePigletFosteringHistoryBySowUseCase{},
			&fakePartialWeagingHistoryBySowUseCase{},
			&fakeWeagingHistoryBySowUseCase{},
		)

		response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/c-001/history", nil))

		if response.Code != http.StatusOK {
			t.Fatalf("status=%d", response.Code)
		}
		body := response.Body.String()
		if !strings.Contains(body, `"events":[{"type":"sow_entry"`) {
			t.Fatalf("expected only the sow entry event: %s", body)
		}
		if strings.Contains(body, `"type":"service"`) || strings.Contains(body, `"type":"abortion"`) ||
			strings.Contains(body, `"type":"farrowing"`) || strings.Contains(body, `"type":"piglet_death"`) ||
			strings.Contains(body, `"type":"piglet_fostering"`) || strings.Contains(body, `"type":"partial_weaging"`) ||
			strings.Contains(body, `"type":"weaging"`) {
			t.Fatalf("unexpected reproductive events: %s", body)
		}
	})

	t.Run("returns not found for missing or inactive sow", func(t *testing.T) {
		for _, testErr := range []error{ports.ErrSowNotFound, sowapplication.ErrSowInactive} {
			serviceUseCase := &fakeServicesHistoryBySowUseCase{}
			abortionUseCase := &fakeAbortionHistoryBySowUseCase{}
			farrowingUseCase := &fakeFarrowingHistoryBySowUseCase{}
			pigletDeathUseCase := &fakePigletDeathHistoryBySowUseCase{}
			pigletFosteringUseCase := &fakePigletFosteringHistoryBySowUseCase{}
			partialWeagingUseCase := &fakePartialWeagingHistoryBySowUseCase{}
			weagingUseCase := &fakeWeagingHistoryBySowUseCase{}
			sowUseCase := &fakeSowHistoryBySowUseCase{err: testErr}
			handler := newHistoryHandler(
				sowUseCase, serviceUseCase, abortionUseCase, farrowingUseCase,
				pigletDeathUseCase, pigletFosteringUseCase, partialWeagingUseCase, weagingUseCase,
			)

			response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/C-001/history", nil))

			if response.Code != http.StatusNotFound {
				t.Fatalf("status=%d", response.Code)
			}
			if serviceUseCase.called || abortionUseCase.called || farrowingUseCase.called ||
				pigletDeathUseCase.called || pigletFosteringUseCase.called || partialWeagingUseCase.called ||
				weagingUseCase.called {
				t.Fatal("history use cases must not be called when the sow fails")
			}
		}
	})

	t.Run("returns internal error for dependency failure", func(t *testing.T) {
		sowFailure := &fakeSowHistoryBySowUseCase{err: errors.New("db down")}
		handler := newHistoryHandler(
			sowFailure,
			&fakeServicesHistoryBySowUseCase{},
			&fakeAbortionHistoryBySowUseCase{},
			&fakeFarrowingHistoryBySowUseCase{},
			&fakePigletDeathHistoryBySowUseCase{},
			&fakePigletFosteringHistoryBySowUseCase{},
			&fakePartialWeagingHistoryBySowUseCase{},
			&fakeWeagingHistoryBySowUseCase{},
		)
		if response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/C-001/history", nil)); response.Code != http.StatusInternalServerError {
			t.Fatalf("sow failure status=%d", response.Code)
		}

		serviceFailure := &fakeServicesHistoryBySowUseCase{err: errors.New("db down")}
		handler = newHistoryHandler(
			&fakeSowHistoryBySowUseCase{history: historySow()},
			serviceFailure,
			&fakeAbortionHistoryBySowUseCase{},
			&fakeFarrowingHistoryBySowUseCase{},
			&fakePigletDeathHistoryBySowUseCase{},
			&fakePigletFosteringHistoryBySowUseCase{},
			&fakePartialWeagingHistoryBySowUseCase{},
			&fakeWeagingHistoryBySowUseCase{},
		)
		if response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/C-001/history", nil)); response.Code != http.StatusInternalServerError {
			t.Fatalf("service failure status=%d", response.Code)
		}

		abortionFailure := &fakeAbortionHistoryBySowUseCase{err: errors.New("db down")}
		handler = newHistoryHandler(
			&fakeSowHistoryBySowUseCase{history: historySow()},
			&fakeServicesHistoryBySowUseCase{},
			abortionFailure,
			&fakeFarrowingHistoryBySowUseCase{},
			&fakePigletDeathHistoryBySowUseCase{},
			&fakePigletFosteringHistoryBySowUseCase{},
			&fakePartialWeagingHistoryBySowUseCase{},
			&fakeWeagingHistoryBySowUseCase{},
		)
		if response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/C-001/history", nil)); response.Code != http.StatusInternalServerError {
			t.Fatalf("abortion failure status=%d", response.Code)
		}

		farrowingFailure := &fakeFarrowingHistoryBySowUseCase{err: errors.New("db down")}
		handler = newHistoryHandler(
			&fakeSowHistoryBySowUseCase{history: historySow()},
			&fakeServicesHistoryBySowUseCase{},
			&fakeAbortionHistoryBySowUseCase{},
			farrowingFailure,
			&fakePigletDeathHistoryBySowUseCase{},
			&fakePigletFosteringHistoryBySowUseCase{},
			&fakePartialWeagingHistoryBySowUseCase{},
			&fakeWeagingHistoryBySowUseCase{},
		)
		if response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/C-001/history", nil)); response.Code != http.StatusInternalServerError {
			t.Fatalf("farrowing failure status=%d", response.Code)
		}

		pigletDeathFailure := &fakePigletDeathHistoryBySowUseCase{err: errors.New("db down")}
		handler = newHistoryHandler(
			&fakeSowHistoryBySowUseCase{history: historySow()},
			&fakeServicesHistoryBySowUseCase{},
			&fakeAbortionHistoryBySowUseCase{},
			&fakeFarrowingHistoryBySowUseCase{},
			pigletDeathFailure,
			&fakePigletFosteringHistoryBySowUseCase{},
			&fakePartialWeagingHistoryBySowUseCase{},
			&fakeWeagingHistoryBySowUseCase{},
		)
		if response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/C-001/history", nil)); response.Code != http.StatusInternalServerError {
			t.Fatalf("piglet death failure status=%d", response.Code)
		}

		pigletFosteringFailure := &fakePigletFosteringHistoryBySowUseCase{err: errors.New("db down")}
		handler = newHistoryHandler(
			&fakeSowHistoryBySowUseCase{history: historySow()},
			&fakeServicesHistoryBySowUseCase{},
			&fakeAbortionHistoryBySowUseCase{},
			&fakeFarrowingHistoryBySowUseCase{},
			&fakePigletDeathHistoryBySowUseCase{},
			pigletFosteringFailure,
			&fakePartialWeagingHistoryBySowUseCase{},
			&fakeWeagingHistoryBySowUseCase{},
		)
		if response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/C-001/history", nil)); response.Code != http.StatusInternalServerError {
			t.Fatalf("piglet fostering failure status=%d", response.Code)
		}

		partialWeagingFailure := &fakePartialWeagingHistoryBySowUseCase{err: errors.New("db down")}
		handler = newHistoryHandler(
			&fakeSowHistoryBySowUseCase{history: historySow()},
			&fakeServicesHistoryBySowUseCase{},
			&fakeAbortionHistoryBySowUseCase{},
			&fakeFarrowingHistoryBySowUseCase{},
			&fakePigletDeathHistoryBySowUseCase{},
			&fakePigletFosteringHistoryBySowUseCase{},
			partialWeagingFailure,
			&fakeWeagingHistoryBySowUseCase{},
		)
		if response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/C-001/history", nil)); response.Code != http.StatusInternalServerError {
			t.Fatalf("partial weaging failure status=%d", response.Code)
		}

		weagingFailure := &fakeWeagingHistoryBySowUseCase{err: errors.New("db down")}
		handler = newHistoryHandler(
			&fakeSowHistoryBySowUseCase{history: historySow()},
			&fakeServicesHistoryBySowUseCase{},
			&fakeAbortionHistoryBySowUseCase{},
			&fakeFarrowingHistoryBySowUseCase{},
			&fakePigletDeathHistoryBySowUseCase{},
			&fakePigletFosteringHistoryBySowUseCase{},
			&fakePartialWeagingHistoryBySowUseCase{},
			weagingFailure,
		)
		if response := serveHistory(handler, httptest.NewRequest(http.MethodGet, "/api/v1/sows/C-001/history", nil)); response.Code != http.StatusInternalServerError {
			t.Fatalf("weaging failure status=%d", response.Code)
		}
	})
}
