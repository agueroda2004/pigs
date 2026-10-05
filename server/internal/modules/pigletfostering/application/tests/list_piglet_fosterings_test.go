package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	pigletfosteringapplication "server/internal/modules/pigletfostering/application"
	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
	"server/internal/modules/pigletfostering/ports"
)

func TestListPigletFosteringsExecute(t *testing.T) {
	t.Run("returns the fosterings from the repository", func(t *testing.T) {
		expected := []*pigletfosteringdomain.PigletFostering{{ID: uuid.New()}}
		repository := &fakePigletFosteringRepository{listFosterings: expected}
		useCase := pigletfosteringapplication.NewListPigletFosteringsService(repository)

		fosterings, err := useCase.Execute(context.Background(), ports.PigletFosteringFilter{})

		if err != nil || len(fosterings) != 1 || fosterings[0] != expected[0] {
			t.Fatalf("unexpected result: err=%v fosterings=%#v", err, fosterings)
		}
	})

	t.Run("forwards the sow and date filters to the repository", func(t *testing.T) {
		donorSowID := uuid.New()
		receiverSowID := uuid.New()
		from := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)
		repository := &fakePigletFosteringRepository{}
		useCase := pigletfosteringapplication.NewListPigletFosteringsService(repository)

		_, err := useCase.Execute(context.Background(), ports.PigletFosteringFilter{
			DonorSowID:    &donorSowID,
			ReceiverSowID: &receiverSowID,
			FromDate:      &from,
			ToDate:        &to,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.listFilter.DonorSowID == nil || *repository.listFilter.DonorSowID != donorSowID {
			t.Fatalf("unexpected donor filter: %#v", repository.listFilter)
		}
		if repository.listFilter.ReceiverSowID == nil || *repository.listFilter.ReceiverSowID != receiverSowID {
			t.Fatalf("unexpected receiver filter: %#v", repository.listFilter)
		}
		if repository.listFilter.FromDate == nil || !repository.listFilter.FromDate.Equal(from) {
			t.Fatalf("unexpected from filter: %#v", repository.listFilter.FromDate)
		}
		if repository.listFilter.ToDate == nil || !repository.listFilter.ToDate.Equal(to) {
			t.Fatalf("unexpected to filter: %#v", repository.listFilter.ToDate)
		}
	})

	t.Run("propagates a repository error", func(t *testing.T) {
		repository := &fakePigletFosteringRepository{listErr: errors.New("boom")}
		useCase := pigletfosteringapplication.NewListPigletFosteringsService(repository)

		_, err := useCase.Execute(context.Background(), ports.PigletFosteringFilter{})

		if !errors.Is(err, repository.listErr) {
			t.Fatalf("error = %v, want the repository error", err)
		}
	})
}
