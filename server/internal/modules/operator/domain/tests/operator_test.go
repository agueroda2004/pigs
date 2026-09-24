package tests

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	operatordomain "server/internal/modules/operator/domain"
)

func TestNewOperator(t *testing.T) {
	now := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	operatorID := uuid.New()
	createdBy := uuid.New()

	t.Run("creates a valid active operator", func(t *testing.T) {
		operator, err := operatordomain.NewOperator(operatorID, "  Juan Pérez  ", createdBy, now)

		if err != nil {
			t.Fatalf("NewOperator() error = %v", err)
		}
		if operator.ID != operatorID || operator.Name != "Juan Pérez" || !operator.Active {
			t.Fatalf("unexpected operator: %#v", operator)
		}
		if operator.CreatedBy != createdBy || operator.UpdatedBy != createdBy {
			t.Fatalf("unexpected audit fields: %#v", operator)
		}
		if !operator.CreatedAt.Equal(now) || !operator.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected timestamps: %#v", operator)
		}
	})

	t.Run("rejects nil id", func(t *testing.T) {
		_, err := operatordomain.NewOperator(uuid.Nil, "Juan Pérez", createdBy, now)
		if !errors.Is(err, operatordomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})

	t.Run("rejects invalid name", func(t *testing.T) {
		_, err := operatordomain.NewOperator(operatorID, "  ", createdBy, now)
		if !errors.Is(err, operatordomain.ErrInvalidName) {
			t.Fatalf("error = %v, want ErrInvalidName", err)
		}

		_, err = operatordomain.NewOperator(operatorID, strings.Repeat("a", 101), createdBy, now)
		if !errors.Is(err, operatordomain.ErrInvalidName) {
			t.Fatalf("error = %v, want ErrInvalidName", err)
		}
	})

	t.Run("accepts name at exactly max length", func(t *testing.T) {
		name := strings.Repeat("a", 100)
		operator, err := operatordomain.NewOperator(operatorID, name, createdBy, now)
		if err != nil || operator.Name != name {
			t.Fatalf("unexpected result: err=%v operator=%#v", err, operator)
		}
	})

	t.Run("rejects nil created by", func(t *testing.T) {
		_, err := operatordomain.NewOperator(operatorID, "Juan Pérez", uuid.Nil, now)
		if !errors.Is(err, operatordomain.ErrInvalidCreatedBy) {
			t.Fatalf("error = %v, want ErrInvalidCreatedBy", err)
		}
	})
}

func TestOperatorUpdate(t *testing.T) {
	operatorID := uuid.New()
	updatedBy := uuid.New()
	now := time.Date(2026, time.February, 3, 4, 5, 6, 0, time.UTC)

	t.Run("updates name and active", func(t *testing.T) {
		operator := &operatordomain.Operator{ID: operatorID, Name: "Old", Active: true}
		name := "  New Name  "
		active := false

		err := operator.Update(&name, &active, updatedBy, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if operator.Name != "New Name" || operator.Active || operator.UpdatedBy != updatedBy || !operator.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected operator: %#v", operator)
		}
	})

	t.Run("keeps omitted fields unchanged", func(t *testing.T) {
		operator := &operatordomain.Operator{ID: operatorID, Name: "Old", Active: true}
		name := "Updated"

		err := operator.Update(&name, nil, updatedBy, now)

		if err != nil || operator.Name != name || !operator.Active {
			t.Fatalf("unexpected result: err=%v operator=%#v", err, operator)
		}
	})

	t.Run("rejects empty update", func(t *testing.T) {
		operator := &operatordomain.Operator{ID: operatorID, Name: "Old", Active: true}

		err := operator.Update(nil, nil, updatedBy, now)

		if !errors.Is(err, operatordomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("rejects nil updated by", func(t *testing.T) {
		operator := &operatordomain.Operator{ID: operatorID, Name: "Old", Active: true}
		name := "New Name"

		err := operator.Update(&name, nil, uuid.Nil, now)

		if !errors.Is(err, operatordomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})

	t.Run("rejects invalid name without mutating the operator", func(t *testing.T) {
		operator := &operatordomain.Operator{ID: operatorID, Name: "Old", Active: true}
		name := "  "
		active := false

		err := operator.Update(&name, &active, updatedBy, now)

		if !errors.Is(err, operatordomain.ErrInvalidName) || operator.Name != "Old" || !operator.Active {
			t.Fatalf("unexpected result: err=%v operator=%#v", err, operator)
		}
	})

	t.Run("rejects nil receiver", func(t *testing.T) {
		var operator *operatordomain.Operator
		name := "New Name"

		err := operator.Update(&name, nil, updatedBy, now)

		if !errors.Is(err, operatordomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})
}
