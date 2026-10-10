package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	operatorapplication "server/internal/modules/operator/application"
	operatordomain "server/internal/modules/operator/domain"
)

func TestListOperatorDropdownServiceExecute(t *testing.T) {
	t.Run("returns the dropdown items when active is true", func(t *testing.T) {
		options := []operatordomain.OperatorDropdown{
			{ID: uuid.New(), Name: "Ana", Active: true},
			{ID: uuid.New(), Name: "Beto", Active: true},
		}
		repository := &fakeOperatorRepository{listDropdown: options}
		service := operatorapplication.NewListOperatorDropdownService(repository)
		active := true

		result, err := service.Execute(context.Background(), &active)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(result) != len(options) || result[0] != options[0] || result[1] != options[1] {
			t.Fatalf("unexpected options: %#v", result)
		}
		if repository.listDropdownActive == nil || !*repository.listDropdownActive {
			t.Fatalf("unexpected active filter: %#v", repository.listDropdownActive)
		}
	})

	t.Run("forwards a false active filter", func(t *testing.T) {
		repository := &fakeOperatorRepository{listDropdown: []operatordomain.OperatorDropdown{}}
		service := operatorapplication.NewListOperatorDropdownService(repository)
		inactive := false

		result, err := service.Execute(context.Background(), &inactive)

		if err != nil || len(result) != 0 {
			t.Fatalf("unexpected result: options=%#v err=%v", result, err)
		}
		if repository.listDropdownActive == nil || *repository.listDropdownActive {
			t.Fatalf("unexpected active filter: %#v", repository.listDropdownActive)
		}
	})

	t.Run("forwards a nil active filter", func(t *testing.T) {
		repository := &fakeOperatorRepository{listDropdown: []operatordomain.OperatorDropdown{}}
		service := operatorapplication.NewListOperatorDropdownService(repository)

		_, err := service.Execute(context.Background(), nil)

		if err != nil || repository.listDropdownActive != nil {
			t.Fatalf("unexpected result: active=%#v err=%v", repository.listDropdownActive, err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeOperatorRepository{listDropdownErr: expected}
		service := operatorapplication.NewListOperatorDropdownService(repository)

		_, err := service.Execute(context.Background(), nil)

		if !errors.Is(err, expected) {
			t.Fatalf("Execute() error = %v, want %v", err, expected)
		}
	})
}
