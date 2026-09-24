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

func TestUpdateOperatorServiceExecute(t *testing.T) {
	operatorID := uuid.New()
	updatedBy := uuid.New()
	now := time.Date(2026, time.March, 4, 5, 6, 7, 0, time.UTC)

	t.Run("updates name and active", func(t *testing.T) {
		operator := testOperator(operatorID)
		name := "  New Name  "
		active := false
		repository := &fakeOperatorRepository{getOperator: operator}
		service := operatorapplication.NewUpdateOperatorService(repository, func() time.Time { return now })

		updated, err := service.Execute(context.Background(), operatorID, operatorapplication.UpdateOperatorCommand{
			Name: &name, Active: &active, UpdatedBy: updatedBy,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if updated != operator || operator.Name != "New Name" || operator.Active || operator.UpdatedBy != updatedBy || !operator.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected operator: %#v", operator)
		}
		if repository.updated != operator {
			t.Fatalf("repository did not receive the updated operator")
		}
	})

	t.Run("keeps omitted fields unchanged", func(t *testing.T) {
		operator := testOperator(operatorID)
		active := false
		repository := &fakeOperatorRepository{getOperator: operator}
		service := operatorapplication.NewUpdateOperatorService(repository, time.Now)

		_, err := service.Execute(context.Background(), operatorID, operatorapplication.UpdateOperatorCommand{Active: &active, UpdatedBy: updatedBy})

		if err != nil || operator.Active || operator.Name != "Old Name" {
			t.Fatalf("unexpected result: err=%v operator=%#v", err, operator)
		}
	})

	t.Run("propagates dependency and validation errors", func(t *testing.T) {
		getErr := errors.New("get failed")
		repository := &fakeOperatorRepository{getErr: getErr}
		service := operatorapplication.NewUpdateOperatorService(repository, time.Now)
		_, err := service.Execute(context.Background(), operatorID, operatorapplication.UpdateOperatorCommand{UpdatedBy: updatedBy})
		if !errors.Is(err, getErr) {
			t.Fatalf("get error = %v", err)
		}

		repository = &fakeOperatorRepository{getOperator: testOperator(operatorID)}
		service = operatorapplication.NewUpdateOperatorService(repository, time.Now)
		_, err = service.Execute(context.Background(), operatorID, operatorapplication.UpdateOperatorCommand{UpdatedBy: updatedBy})
		if !errors.Is(err, operatordomain.ErrInvalidUpdate) || repository.updated != nil {
			t.Fatalf("empty command error = %v", err)
		}

		name := "New Name"
		repository = &fakeOperatorRepository{getOperator: testOperator(operatorID), updateErr: ports.ErrOperatorNameAlreadyUsed}
		service = operatorapplication.NewUpdateOperatorService(repository, time.Now)
		_, err = service.Execute(context.Background(), operatorID, operatorapplication.UpdateOperatorCommand{Name: &name, UpdatedBy: updatedBy})
		if !errors.Is(err, ports.ErrOperatorNameAlreadyUsed) {
			t.Fatalf("repository error = %v", err)
		}
	})
}
