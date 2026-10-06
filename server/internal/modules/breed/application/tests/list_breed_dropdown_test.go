package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	breedapplication "server/internal/modules/breed/application"
	breeddomain "server/internal/modules/breed/domain"
)

func TestListBreedDropdownServiceExecute(t *testing.T) {
	t.Run("returns every breed when active is false", func(t *testing.T) {
		options := []breeddomain.BreedDropdown{
			{ID: uuid.New(), Name: "Duroc", Active: true},
			{ID: uuid.New(), Name: "Landrace", Active: false},
		}
		repository := &fakeBreedRepository{listDropdown: options}
		service := breedapplication.NewListBreedDropdownService(repository)

		result, err := service.Execute(context.Background(), false)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.listDropdownActive {
			t.Fatalf("expected active filter false")
		}
		if len(result) != len(options) || result[0] != options[0] || result[1] != options[1] {
			t.Fatalf("unexpected options: %#v", result)
		}
	})

	t.Run("forwards active true to only return active breeds", func(t *testing.T) {
		options := []breeddomain.BreedDropdown{{ID: uuid.New(), Name: "Duroc", Active: true}}
		repository := &fakeBreedRepository{listDropdown: options}
		service := breedapplication.NewListBreedDropdownService(repository)

		result, err := service.Execute(context.Background(), true)

		if err != nil || !repository.listDropdownActive || len(result) != 1 {
			t.Fatalf("unexpected result: active=%v options=%#v err=%v", repository.listDropdownActive, result, err)
		}
	})

	t.Run("returns an empty list when there are no breeds", func(t *testing.T) {
		repository := &fakeBreedRepository{listDropdown: []breeddomain.BreedDropdown{}}
		service := breedapplication.NewListBreedDropdownService(repository)

		result, err := service.Execute(context.Background(), false)

		if err != nil || len(result) != 0 {
			t.Fatalf("unexpected result: options=%#v err=%v", result, err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeBreedRepository{listDropdownErr: expected}
		service := breedapplication.NewListBreedDropdownService(repository)

		_, err := service.Execute(context.Background(), false)

		if !errors.Is(err, expected) {
			t.Fatalf("Execute() error = %v, want %v", err, expected)
		}
	})
}
