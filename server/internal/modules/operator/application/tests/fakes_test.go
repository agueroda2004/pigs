package tests

import (
	"context"

	"github.com/google/uuid"

	operatordomain "server/internal/modules/operator/domain"
)

type fakeOperatorRepository struct {
	exists        bool
	existsErr     error
	getOperator   *operatordomain.Operator
	getErr        error
	listOperators []*operatordomain.Operator
	listErr       error
	createErr     error
	updateErr     error
	created       *operatordomain.Operator
	updated       *operatordomain.Operator
}

func (f *fakeOperatorRepository) Create(_ context.Context, operator *operatordomain.Operator) error {
	f.created = operator
	return f.createErr
}

func (f *fakeOperatorRepository) GetByID(_ context.Context, _ uuid.UUID) (*operatordomain.Operator, error) {
	return f.getOperator, f.getErr
}

func (f *fakeOperatorRepository) ExistsByName(_ context.Context, _ string) (bool, error) {
	return f.exists, f.existsErr
}

func (f *fakeOperatorRepository) List(_ context.Context) ([]*operatordomain.Operator, error) {
	return f.listOperators, f.listErr
}

func (f *fakeOperatorRepository) Update(_ context.Context, operator *operatordomain.Operator) error {
	f.updated = operator
	return f.updateErr
}
