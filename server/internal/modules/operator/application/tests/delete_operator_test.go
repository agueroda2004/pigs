package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	operatorapplication "server/internal/modules/operator/application"
	"server/internal/modules/operator/ports"
)

func TestDeleteOperatorServiceExecute(t *testing.T) {
	operatorID := uuid.New()

	t.Run("deletes the operator by its identifier", func(t *testing.T) {
		repository := &fakeOperatorRepository{}
		service := operatorapplication.NewDeleteOperatorService(repository)

		err := service.Execute(context.Background(), operatorID)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !repository.deleteCalled || repository.deletedID != operatorID {
			t.Fatalf("repository did not receive the operator id: %#v", repository)
		}
	})

	t.Run("returns not found when the operator does not exist", func(t *testing.T) {
		repository := &fakeOperatorRepository{deleteErr: ports.ErrOperatorNotFound}
		service := operatorapplication.NewDeleteOperatorService(repository)

		err := service.Execute(context.Background(), operatorID)

		if !errors.Is(err, ports.ErrOperatorNotFound) {
			t.Fatalf("Execute() error = %v, want ErrOperatorNotFound", err)
		}
	})

	t.Run("returns in use when the operator has related records", func(t *testing.T) {
		repository := &fakeOperatorRepository{deleteErr: ports.ErrOperatorInUse}
		service := operatorapplication.NewDeleteOperatorService(repository)

		err := service.Execute(context.Background(), operatorID)

		if !errors.Is(err, ports.ErrOperatorInUse) {
			t.Fatalf("Execute() error = %v, want ErrOperatorInUse", err)
		}
	})
}
