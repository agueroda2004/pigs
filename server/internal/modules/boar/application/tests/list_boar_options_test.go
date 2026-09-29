package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	boarapplication "server/internal/modules/boar/application"
	boardomain "server/internal/modules/boar/domain"
)

func TestListBoarOptionsServiceExecute(t *testing.T) {
	t.Run("returns the alive boar options when active is true", func(t *testing.T) {
		options := []boardomain.BoarOption{
			{ID: uuid.New(), Code: "B-001"},
			{ID: uuid.New(), Code: "B-002"},
		}
		repository := &fakeBoarRepository{listOptions: options}
		service := boarapplication.NewListBoarOptionsService(repository)
		active := true

		result, err := service.Execute(context.Background(), &active)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(result) != len(options) || result[0] != options[0] || result[1] != options[1] {
			t.Fatalf("unexpected options: %#v", result)
		}
		if repository.listOptionsActive == nil || *repository.listOptionsActive != active {
			t.Fatalf("unexpected active filter: %#v", repository.listOptionsActive)
		}
	})

	t.Run("forwards a nil active filter to list every boar", func(t *testing.T) {
		repository := &fakeBoarRepository{listOptions: []boardomain.BoarOption{}}
		service := boarapplication.NewListBoarOptionsService(repository)

		result, err := service.Execute(context.Background(), nil)

		if err != nil || len(result) != 0 {
			t.Fatalf("unexpected result: options=%#v err=%v", result, err)
		}
		if repository.listOptionsActive != nil {
			t.Fatalf("unexpected active filter: %#v", repository.listOptionsActive)
		}
	})

	t.Run("forwards a false active filter", func(t *testing.T) {
		repository := &fakeBoarRepository{listOptions: []boardomain.BoarOption{}}
		service := boarapplication.NewListBoarOptionsService(repository)
		active := false

		result, err := service.Execute(context.Background(), &active)

		if err != nil || len(result) != 0 {
			t.Fatalf("unexpected result: options=%#v err=%v", result, err)
		}
		if repository.listOptionsActive == nil || *repository.listOptionsActive {
			t.Fatalf("unexpected active filter: %#v", repository.listOptionsActive)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeBoarRepository{listOptionsErr: expected}
		service := boarapplication.NewListBoarOptionsService(repository)

		_, err := service.Execute(context.Background(), nil)

		if !errors.Is(err, expected) {
			t.Fatalf("Execute() error = %v, want %v", err, expected)
		}
	})
}
