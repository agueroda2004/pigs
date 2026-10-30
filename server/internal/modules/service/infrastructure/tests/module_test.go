package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	servicedomain "server/internal/modules/service/domain"
	serviceinfra "server/internal/modules/service/infrastructure"
	"server/internal/modules/service/ports"
)

// stubScheduleRepository keeps the full repository contract through embedding.
// Tests only exercise the schedule query, so the remaining methods are unused.
type stubScheduleRepository struct {
	ports.ServiceRepository
	schedule  *servicedomain.FarrowingSchedule
	err       error
	reference time.Time
	called    bool
}

func (s *stubScheduleRepository) GetFarrowingSchedule(_ context.Context, reference time.Time) (*servicedomain.FarrowingSchedule, error) {
	s.called = true
	s.reference = reference
	return s.schedule, s.err
}

func TestNewModuleWiresFarrowingSchedule(t *testing.T) {
	now := time.Date(2026, time.June, 15, 8, 0, 0, 0, time.UTC)
	average := 112.0
	repository := &stubScheduleRepository{schedule: &servicedomain.FarrowingSchedule{
		Upcoming: []servicedomain.FarrowingScheduleEntry{{
			ServiceID:             uuid.New(),
			SowID:                 uuid.New(),
			SowCode:               "C-001",
			ExpectedFarrowingDate: time.Date(2026, time.June, 20, 0, 0, 0, 0, time.UTC),
			AverageGestationDays:  &average,
		}},
		Overdue: []servicedomain.FarrowingScheduleEntry{},
	}}
	passThrough := func(next http.Handler) http.Handler { return next }
	module := serviceinfra.NewModule(repository, func() time.Time { return now }, passThrough, passThrough)

	if module.FarrowingSchedule == nil {
		t.Fatal("expected the module to expose the farrowing schedule use case")
	}

	mux := http.NewServeMux()
	module.Handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/services/farrowing-schedule", nil))

	if response.Code != http.StatusOK || !repository.called {
		t.Fatalf("status=%d called=%v", response.Code, repository.called)
	}
	if !repository.reference.Equal(now) {
		t.Fatalf("reference = %v, want %v", repository.reference, now)
	}
	body := response.Body.String()
	if !strings.Contains(body, `"sow_code":"C-001"`) || !strings.Contains(body, `"average_gestation_days":112`) {
		t.Fatalf("unexpected body: %s", body)
	}
}
