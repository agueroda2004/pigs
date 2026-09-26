package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	sowapplication "server/internal/modules/sow/application"
	sowdomain "server/internal/modules/sow/domain"
)

func TestListSowOptionsServiceExecute(t *testing.T) {
	t.Run("returns the serviceable sow options when active is true", func(t *testing.T) {
		options := []sowdomain.SowOption{
			{ID: uuid.New(), Code: "C-001"},
			{ID: uuid.New(), Code: "C-002"},
		}
		repository := &fakeSowRepository{listOptions: options}
		service := sowapplication.NewListSowOptionsService(repository)
		active := true

		result, err := service.Execute(context.Background(), &active)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(result) != len(options) || result[0] != options[0] || result[1] != options[1] {
			t.Fatalf("unexpected options: %#v", result)
		}
		if repository.listActive == nil || *repository.listActive != active {
			t.Fatalf("unexpected active filter: %#v", repository.listActive)
		}
	})

	t.Run("forwards a nil active filter to list every sow", func(t *testing.T) {
		repository := &fakeSowRepository{listOptions: []sowdomain.SowOption{}}
		service := sowapplication.NewListSowOptionsService(repository)

		result, err := service.Execute(context.Background(), nil)

		if err != nil || len(result) != 0 {
			t.Fatalf("unexpected result: options=%#v err=%v", result, err)
		}
		if repository.listActive != nil {
			t.Fatalf("unexpected active filter: %#v", repository.listActive)
		}
	})

	t.Run("forwards a false active filter", func(t *testing.T) {
		repository := &fakeSowRepository{listOptions: []sowdomain.SowOption{}}
		service := sowapplication.NewListSowOptionsService(repository)
		active := false

		result, err := service.Execute(context.Background(), &active)

		if err != nil || len(result) != 0 {
			t.Fatalf("unexpected result: options=%#v err=%v", result, err)
		}
		if repository.listActive == nil || *repository.listActive {
			t.Fatalf("unexpected active filter: %#v", repository.listActive)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeSowRepository{listOptionsErr: expected}
		service := sowapplication.NewListSowOptionsService(repository)

		_, err := service.Execute(context.Background(), nil)

		if !errors.Is(err, expected) {
			t.Fatalf("Execute() error = %v, want %v", err, expected)
		}
	})
}
