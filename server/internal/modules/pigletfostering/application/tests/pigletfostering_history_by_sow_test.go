package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	pigletfosteringapplication "server/internal/modules/pigletfostering/application"
	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
)

func TestPigletFosteringHistoryBySowExecute(t *testing.T) {
	t.Run("returns the fostering history of the requested sow", func(t *testing.T) {
		note := "Traslado a la cerda 106"
		sowID := uuid.New()
		items := []*pigletfosteringdomain.PigletFosteringHistoryBySow{
			{
				SowID:        sowID,
				Role:         pigletfosteringdomain.PigletFosteringRoleDonor,
				MovementDate: time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC),
				Quantity:     -2,
				Note:         &note,
			},
		}
		repository := &fakePigletFosteringRepository{historyItems: items}
		useCase := pigletfosteringapplication.NewPigletFosteringHistoryBySowService(repository)

		result, err := useCase.Execute(context.Background(), "C-001")

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.historyCode != "C-001" {
			t.Fatalf("sow code = %q, want C-001", repository.historyCode)
		}
		if len(result) != 1 || result[0].SowID != sowID ||
			result[0].Role != pigletfosteringdomain.PigletFosteringRoleDonor ||
			result[0].Quantity != -2 {
			t.Fatalf("unexpected result: %#v", result)
		}
		if result[0].Note == nil || *result[0].Note != note {
			t.Fatalf("note = %v, want %q", result[0].Note, note)
		}
	})

	t.Run("returns an empty slice when there is no history", func(t *testing.T) {
		repository := &fakePigletFosteringRepository{}
		useCase := pigletfosteringapplication.NewPigletFosteringHistoryBySowService(repository)

		result, err := useCase.Execute(context.Background(), "C-001")

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result == nil || len(result) != 0 {
			t.Fatalf("result = %#v, want an empty slice", result)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := &fakePigletFosteringRepository{historyErr: errors.New("boom")}
		useCase := pigletfosteringapplication.NewPigletFosteringHistoryBySowService(repository)

		_, err := useCase.Execute(context.Background(), "C-001")

		if !errors.Is(err, repository.historyErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
