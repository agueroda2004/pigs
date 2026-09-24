package tests

import (
	"time"

	"github.com/google/uuid"

	operatordomain "server/internal/modules/operator/domain"
)

func testOperator(id uuid.UUID) *operatordomain.Operator {
	return &operatordomain.Operator{
		ID: id, Name: "Old Name", Active: true,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
}
