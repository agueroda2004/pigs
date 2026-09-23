package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	breedapplication "server/internal/modules/breed/application"
	breeddomain "server/internal/modules/breed/domain"
)

func TestListBreedsServiceExecute(t *testing.T) {
	t.Run("returns every breed from the repository", func(t *testing.T) {
		breeds := []*breeddomain.Breed{
			testBreed(uuid.New()),
			testBreed(uuid.New()),
		}
		repository := &fakeBreedRepository{listBreeds: breeds}
		service := breedapplication.NewListBreedsService(repository)

		result, err := service.Execute(context.Background())

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(result) != len(breeds) || result[0] != breeds[0] || result[1] != breeds[1] {
			t.Fatalf("unexpected breeds: %#v", result)
		}
	})

	t.Run("returns an empty list when there are no breeds", func(t *testing.T) {
		repository := &fakeBreedRepository{listBreeds: []*breeddomain.Breed{}}
		service := breedapplication.NewListBreedsService(repository)

		result, err := service.Execute(context.Background())

		if err != nil || len(result) != 0 {
			t.Fatalf("unexpected result: breeds=%#v err=%v", result, err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeBreedRepository{listErr: expected}
		service := breedapplication.NewListBreedsService(repository)

		_, err := service.Execute(context.Background())

		if !errors.Is(err, expected) {
			t.Fatalf("Execute() error = %v, want %v", err, expected)
		}
	})
}
