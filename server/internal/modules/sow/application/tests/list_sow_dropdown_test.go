package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	sowapplication "server/internal/modules/sow/application"
	sowdomain "server/internal/modules/sow/domain"
)

func TestListSowDropdownServiceExecute(t *testing.T) {
	t.Run("returns the dropdown items matching active and states", func(t *testing.T) {
		options := []sowdomain.SowDropdown{
			{ID: uuid.New(), Code: "C-001"},
			{ID: uuid.New(), Code: "C-002"},
		}
		repository := &fakeSowRepository{listDropdown: options}
		service := sowapplication.NewListSowDropdownService(repository)
		active := true
		states := []sowdomain.State{sowdomain.StateAlive, sowdomain.StateWeaned}

		result, err := service.Execute(context.Background(), &active, states)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(result) != len(options) || result[0] != options[0] || result[1] != options[1] {
			t.Fatalf("unexpected options: %#v", result)
		}
		if repository.listActive == nil || *repository.listActive != active {
			t.Fatalf("unexpected active filter: %#v", repository.listActive)
		}
		if len(repository.listStates) != 2 || repository.listStates[0] != sowdomain.StateAlive {
			t.Fatalf("unexpected states filter: %#v", repository.listStates)
		}
	})

	t.Run("forwards a nil active filter and no states", func(t *testing.T) {
		repository := &fakeSowRepository{listDropdown: []sowdomain.SowDropdown{}}
		service := sowapplication.NewListSowDropdownService(repository)

		result, err := service.Execute(context.Background(), nil, nil)

		if err != nil || len(result) != 0 {
			t.Fatalf("unexpected result: options=%#v err=%v", result, err)
		}
		if repository.listActive != nil || repository.listStates != nil {
			t.Fatalf("unexpected filters: active=%#v states=%#v", repository.listActive, repository.listStates)
		}
	})

	t.Run("forwards a false active filter", func(t *testing.T) {
		repository := &fakeSowRepository{listDropdown: []sowdomain.SowDropdown{}}
		service := sowapplication.NewListSowDropdownService(repository)
		active := false

		result, err := service.Execute(context.Background(), &active, nil)

		if err != nil || len(result) != 0 {
			t.Fatalf("unexpected result: options=%#v err=%v", result, err)
		}
		if repository.listActive == nil || *repository.listActive {
			t.Fatalf("unexpected active filter: %#v", repository.listActive)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeSowRepository{listDropdownErr: expected}
		service := sowapplication.NewListSowDropdownService(repository)

		_, err := service.Execute(context.Background(), nil, nil)

		if !errors.Is(err, expected) {
			t.Fatalf("Execute() error = %v, want %v", err, expected)
		}
	})
}
