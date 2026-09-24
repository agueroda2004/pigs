package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	operatorapplication "server/internal/modules/operator/application"
	operatordomain "server/internal/modules/operator/domain"
	"server/internal/modules/operator/ports"
)

func TestCreateOperatorServiceExecute(t *testing.T) {
	createdAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	createdBy := uuid.New()

	t.Run("creates an operator with a normalized name", func(t *testing.T) {
		repository := &fakeOperatorRepository{}
		service := operatorapplication.NewCreateOperatorService(repository, func() time.Time { return createdAt })

		operator, err := service.Execute(context.Background(), operatorapplication.CreateOperatorCommand{
			Name:      "  Juan Pérez  ",
			CreatedBy: createdBy,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if operator.ID == uuid.Nil || operator.Name != "Juan Pérez" || !operator.Active {
			t.Fatalf("unexpected operator: %#v", operator)
		}
		if operator.CreatedBy != createdBy || operator.UpdatedBy != createdBy || !operator.CreatedAt.Equal(createdAt) {
			t.Fatalf("unexpected audit fields: %#v", operator)
		}
		if repository.created != operator {
			t.Fatalf("repository did not receive the created operator")
		}
	})

	t.Run("returns duplicate name error", func(t *testing.T) {
		repository := &fakeOperatorRepository{exists: true}
		service := operatorapplication.NewCreateOperatorService(repository, time.Now)

		_, err := service.Execute(context.Background(), operatorapplication.CreateOperatorCommand{Name: "Juan Pérez", CreatedBy: createdBy})

		if !errors.Is(err, ports.ErrOperatorNameAlreadyUsed) || repository.created != nil {
			t.Fatalf("unexpected result: err=%v created=%v", err, repository.created)
		}
	})

	t.Run("propagates repository existence error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeOperatorRepository{existsErr: expected}
		service := operatorapplication.NewCreateOperatorService(repository, time.Now)

		_, err := service.Execute(context.Background(), operatorapplication.CreateOperatorCommand{Name: "Juan Pérez", CreatedBy: createdBy})

		if !errors.Is(err, expected) {
			t.Fatalf("Execute() error = %v, want %v", err, expected)
		}
	})

	t.Run("propagates validation and create errors", func(t *testing.T) {
		repository := &fakeOperatorRepository{createErr: errors.New("create failed")}
		service := operatorapplication.NewCreateOperatorService(repository, time.Now)

		_, err := service.Execute(context.Background(), operatorapplication.CreateOperatorCommand{Name: "  ", CreatedBy: createdBy})
		if !errors.Is(err, operatordomain.ErrInvalidName) || repository.created != nil {
			t.Fatalf("unexpected validation result: %v", err)
		}

		_, err = service.Execute(context.Background(), operatorapplication.CreateOperatorCommand{Name: "Juan Pérez", CreatedBy: createdBy})
		if !errors.Is(err, repository.createErr) {
			t.Fatalf("unexpected create result: %v", err)
		}
	})
}
