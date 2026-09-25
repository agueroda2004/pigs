package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	serviceapplication "server/internal/modules/service/application"
	servicedomain "server/internal/modules/service/domain"
	"server/internal/modules/service/ports"
)

func testService(id, sowID uuid.UUID) *servicedomain.Service {
	return &servicedomain.Service{
		ID:    id,
		SowID: sowID,
		State: servicedomain.StateConfirmed,
	}
}

func TestListServices(t *testing.T) {
	t.Run("returns every registered service", func(t *testing.T) {
		repository := &fakeServiceRepository{listServices: []*servicedomain.Service{
			testService(uuid.New(), uuid.New()),
			testService(uuid.New(), uuid.New()),
		}}
		service := serviceapplication.NewListServicesService(repository)

		result, err := service.Execute(context.Background(), ports.ServiceFilter{})

		if err != nil || len(result) != 2 {
			t.Fatalf("unexpected result: err=%v result=%#v", err, result)
		}
	})

	t.Run("forwards the filter to the repository", func(t *testing.T) {
		repository := &fakeServiceRepository{}
		service := serviceapplication.NewListServicesService(repository)
		sowID := uuid.New()
		state := servicedomain.StateFailed

		_, err := service.Execute(context.Background(), ports.ServiceFilter{SowID: &sowID, State: &state})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.listFilter.SowID == nil || *repository.listFilter.SowID != sowID {
			t.Fatalf("unexpected sow id: %#v", repository.listFilter.SowID)
		}
		if repository.listFilter.State == nil || *repository.listFilter.State != state {
			t.Fatalf("unexpected state: %#v", repository.listFilter.State)
		}
	})

	t.Run("propagates the list error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeServiceRepository{listErr: unexpected}
		service := serviceapplication.NewListServicesService(repository)

		_, err := service.Execute(context.Background(), ports.ServiceFilter{})

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
	})
}
