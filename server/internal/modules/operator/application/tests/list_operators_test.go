package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	operatorapplication "server/internal/modules/operator/application"
	operatordomain "server/internal/modules/operator/domain"
)

func TestListOperatorsServiceExecute(t *testing.T) {
	t.Run("returns every operator from the repository", func(t *testing.T) {
		operators := []*operatordomain.Operator{
			testOperator(uuid.New()),
			testOperator(uuid.New()),
		}
		repository := &fakeOperatorRepository{listOperators: operators}
		service := operatorapplication.NewListOperatorsService(repository)

		result, err := service.Execute(context.Background())

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(result) != len(operators) || result[0] != operators[0] || result[1] != operators[1] {
			t.Fatalf("unexpected operators: %#v", result)
		}
	})

	t.Run("returns an empty list when there are no operators", func(t *testing.T) {
		repository := &fakeOperatorRepository{listOperators: []*operatordomain.Operator{}}
		service := operatorapplication.NewListOperatorsService(repository)

		result, err := service.Execute(context.Background())

		if err != nil || len(result) != 0 {
			t.Fatalf("unexpected result: operators=%#v err=%v", result, err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeOperatorRepository{listErr: expected}
		service := operatorapplication.NewListOperatorsService(repository)

		_, err := service.Execute(context.Background())

		if !errors.Is(err, expected) {
			t.Fatalf("Execute() error = %v, want %v", err, expected)
		}
	})
}
