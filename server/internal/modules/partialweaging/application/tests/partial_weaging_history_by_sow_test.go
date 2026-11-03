package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	partialweagingapplication "server/internal/modules/partialweaging/application"
	partialweagingdomain "server/internal/modules/partialweaging/domain"
)

func TestPartialWeagingHistoryBySowExecute(t *testing.T) {
	t.Run("returns every partial weaging for the requested sow", func(t *testing.T) {
		weight := 42.5
		note := "Destete parcial de la camada"
		items := []*partialweagingdomain.PartialWeagingHistoryBySow{
			{
				WeagingDate: time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC),
				Quantity:    2,
				TotalWeight: &weight,
				Type:        partialweagingdomain.TypeNodriza,
				Note:        &note,
			},
		}
		repository := &fakePartialWeagingRepository{historyItems: items}
		useCase := partialweagingapplication.NewPartialWeagingHistoryBySowService(repository)

		result, err := useCase.Execute(context.Background(), "C-001")

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(result) != 1 || result[0] != items[0] {
			t.Fatalf("unexpected result: %#v", result)
		}
		if repository.historyCode != "C-001" {
			t.Fatalf("history code = %q, want C-001", repository.historyCode)
		}
	})

	t.Run("returns an empty slice when there are no records", func(t *testing.T) {
		repository := &fakePartialWeagingRepository{historyItems: nil}
		useCase := partialweagingapplication.NewPartialWeagingHistoryBySowService(repository)

		result, err := useCase.Execute(context.Background(), "C-001")

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result == nil {
			t.Fatal("result must not be nil")
		}
		if len(result) != 0 {
			t.Fatalf("unexpected result: %#v", result)
		}
	})

	t.Run("propagates a repository error", func(t *testing.T) {
		repository := &fakePartialWeagingRepository{historyErr: errors.New("boom")}
		useCase := partialweagingapplication.NewPartialWeagingHistoryBySowService(repository)

		result, err := useCase.Execute(context.Background(), "C-001")

		if !errors.Is(err, repository.historyErr) {
			t.Fatalf("error = %v, want the repository error", err)
		}
		if result != nil {
			t.Fatalf("result = %#v, want nil", result)
		}
	})
}
