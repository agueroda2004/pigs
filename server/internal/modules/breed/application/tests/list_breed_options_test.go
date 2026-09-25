package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	breedapplication "server/internal/modules/breed/application"
	breeddomain "server/internal/modules/breed/domain"
)

func TestListBreedOptionsServiceExecute(t *testing.T) {
	t.Run("returns the active breed options from the repository", func(t *testing.T) {
		options := []breeddomain.BreedOption{
			{ID: uuid.New(), Name: "Duroc"},
			{ID: uuid.New(), Name: "Landrace"},
		}
		repository := &fakeBreedRepository{listOptions: options}
		service := breedapplication.NewListBreedOptionsService(repository)

		result, err := service.Execute(context.Background())

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(result) != len(options) || result[0] != options[0] || result[1] != options[1] {
			t.Fatalf("unexpected options: %#v", result)
		}
	})

	t.Run("returns an empty list when there are no active breeds", func(t *testing.T) {
		repository := &fakeBreedRepository{listOptions: []breeddomain.BreedOption{}}
		service := breedapplication.NewListBreedOptionsService(repository)

		result, err := service.Execute(context.Background())

		if err != nil || len(result) != 0 {
			t.Fatalf("unexpected result: options=%#v err=%v", result, err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeBreedRepository{listOptionsErr: expected}
		service := breedapplication.NewListBreedOptionsService(repository)

		_, err := service.Execute(context.Background())

		if !errors.Is(err, expected) {
			t.Fatalf("Execute() error = %v, want %v", err, expected)
		}
	})
}
