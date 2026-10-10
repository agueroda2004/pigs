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
	t.Run("returns the first page with its metadata", func(t *testing.T) {
		repository := &fakeServiceRepository{
			listServices: []*servicedomain.Service{
				testService(uuid.New(), uuid.New()),
				testService(uuid.New(), uuid.New()),
			},
			listTotal: 25,
		}
		service := serviceapplication.NewListServicesService(repository)

		result, err := service.Execute(context.Background(), ports.ServiceFilter{}, 1)

		if err != nil || len(result.Items) != 2 {
			t.Fatalf("unexpected result: err=%v result=%#v", err, result)
		}
		if result.Total != 25 || result.Page != 1 || result.PageSize != 10 || result.TotalPages != 3 {
			t.Fatalf("unexpected metadata: %#v", result)
		}
		if repository.listLimit != 10 || repository.listOffset != 0 {
			t.Fatalf("unexpected pagination: limit=%d offset=%d", repository.listLimit, repository.listOffset)
		}
	})

	t.Run("computes the offset for a later page", func(t *testing.T) {
		repository := &fakeServiceRepository{listServices: []*servicedomain.Service{}, listTotal: 0}
		service := serviceapplication.NewListServicesService(repository)

		result, err := service.Execute(context.Background(), ports.ServiceFilter{}, 3)

		if err != nil || result.Page != 3 || result.TotalPages != 0 {
			t.Fatalf("unexpected result: err=%v result=%#v", err, result)
		}
		if repository.listLimit != 10 || repository.listOffset != 20 {
			t.Fatalf("unexpected pagination: limit=%d offset=%d", repository.listLimit, repository.listOffset)
		}
	})

	t.Run("defaults an invalid page to the first one", func(t *testing.T) {
		repository := &fakeServiceRepository{listServices: []*servicedomain.Service{}}
		service := serviceapplication.NewListServicesService(repository)

		result, err := service.Execute(context.Background(), ports.ServiceFilter{}, 0)

		if err != nil || result.Page != 1 || repository.listOffset != 0 {
			t.Fatalf("unexpected result: err=%v result=%#v offset=%d", err, result, repository.listOffset)
		}
	})

	t.Run("forwards the filter to the repository", func(t *testing.T) {
		repository := &fakeServiceRepository{}
		service := serviceapplication.NewListServicesService(repository)
		sowCode := "C-001"
		state := servicedomain.StateFailed

		_, err := service.Execute(context.Background(), ports.ServiceFilter{SowCode: &sowCode, State: &state}, 1)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.listFilter.SowCode == nil || *repository.listFilter.SowCode != sowCode {
			t.Fatalf("unexpected sow code: %#v", repository.listFilter.SowCode)
		}
		if repository.listFilter.State == nil || *repository.listFilter.State != state {
			t.Fatalf("unexpected state: %#v", repository.listFilter.State)
		}
	})

	t.Run("propagates the list error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeServiceRepository{listErr: unexpected}
		service := serviceapplication.NewListServicesService(repository)

		_, err := service.Execute(context.Background(), ports.ServiceFilter{}, 1)

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
	})
}
