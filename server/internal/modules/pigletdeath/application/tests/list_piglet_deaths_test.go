package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	pigletdeathapplication "server/internal/modules/pigletdeath/application"
	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	"server/internal/modules/pigletdeath/ports"
)

func TestListPigletDeathsExecute(t *testing.T) {
	t.Run("returns the piglet deaths from the repository", func(t *testing.T) {
		expected := []*pigletdeathdomain.PigletDeath{{ID: uuid.New()}}
		repository := &fakePigletDeathRepository{listDeaths: expected}
		useCase := pigletdeathapplication.NewListPigletDeathsService(repository)

		deaths, err := useCase.Execute(context.Background(), ports.PigletDeathFilter{})

		if err != nil || len(deaths) != 1 || deaths[0] != expected[0] {
			t.Fatalf("unexpected result: err=%v deaths=%#v", err, deaths)
		}
	})

	t.Run("forwards the sow and date filters to the repository", func(t *testing.T) {
		sowID := uuid.New()
		from := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)
		repository := &fakePigletDeathRepository{}
		useCase := pigletdeathapplication.NewListPigletDeathsService(repository)

		_, err := useCase.Execute(context.Background(), ports.PigletDeathFilter{SowID: &sowID, FromDate: &from, ToDate: &to})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.listFilter.SowID == nil || *repository.listFilter.SowID != sowID {
			t.Fatalf("unexpected sow filter: %#v", repository.listFilter)
		}
		if repository.listFilter.FromDate == nil || !repository.listFilter.FromDate.Equal(from) {
			t.Fatalf("unexpected from filter: %#v", repository.listFilter.FromDate)
		}
		if repository.listFilter.ToDate == nil || !repository.listFilter.ToDate.Equal(to) {
			t.Fatalf("unexpected to filter: %#v", repository.listFilter.ToDate)
		}
	})

	t.Run("propagates a repository error", func(t *testing.T) {
		repository := &fakePigletDeathRepository{listErr: errors.New("boom")}
		useCase := pigletdeathapplication.NewListPigletDeathsService(repository)

		_, err := useCase.Execute(context.Background(), ports.PigletDeathFilter{})

		if !errors.Is(err, repository.listErr) {
			t.Fatalf("error = %v, want the repository error", err)
		}
	})
}
