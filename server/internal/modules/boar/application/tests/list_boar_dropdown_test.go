package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	boarapplication "server/internal/modules/boar/application"
	boardomain "server/internal/modules/boar/domain"
)

func TestListBoarDropdownServiceExecute(t *testing.T) {
	t.Run("returns the dropdown items when active is true", func(t *testing.T) {
		options := []boardomain.BoarDropdown{
			{ID: uuid.New(), Code: "B-001", Active: true},
			{ID: uuid.New(), Code: "B-002", Active: true},
		}
		repository := &fakeBoarRepository{listDropdown: options}
		service := boarapplication.NewListBoarDropdownService(repository)
		active := true

		result, err := service.Execute(context.Background(), &active, nil)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(result) != len(options) || result[0] != options[0] || result[1] != options[1] {
			t.Fatalf("unexpected options: %#v", result)
		}
		if repository.listDropdownActive == nil || !*repository.listDropdownActive {
			t.Fatalf("unexpected active filter: %#v", repository.listDropdownActive)
		}
		if repository.listDropdownState != nil {
			t.Fatalf("unexpected state filter: %#v", repository.listDropdownState)
		}
	})

	t.Run("forwards a false active filter and a state filter", func(t *testing.T) {
		repository := &fakeBoarRepository{listDropdown: []boardomain.BoarDropdown{}}
		service := boarapplication.NewListBoarDropdownService(repository)
		inactive := false
		state := boardomain.StateAlive

		result, err := service.Execute(context.Background(), &inactive, &state)

		if err != nil || len(result) != 0 {
			t.Fatalf("unexpected result: options=%#v err=%v", result, err)
		}
		if repository.listDropdownActive == nil || *repository.listDropdownActive {
			t.Fatalf("unexpected active filter: %#v", repository.listDropdownActive)
		}
		if repository.listDropdownState == nil || *repository.listDropdownState != state {
			t.Fatalf("unexpected state filter: %#v", repository.listDropdownState)
		}
	})

	t.Run("forwards a nil active filter", func(t *testing.T) {
		repository := &fakeBoarRepository{listDropdown: []boardomain.BoarDropdown{}}
		service := boarapplication.NewListBoarDropdownService(repository)

		_, err := service.Execute(context.Background(), nil, nil)

		if err != nil || repository.listDropdownActive != nil {
			t.Fatalf("unexpected result: active=%#v err=%v", repository.listDropdownActive, err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeBoarRepository{listDropdownErr: expected}
		service := boarapplication.NewListBoarDropdownService(repository)

		_, err := service.Execute(context.Background(), nil, nil)

		if !errors.Is(err, expected) {
			t.Fatalf("Execute() error = %v, want %v", err, expected)
		}
	})
}
