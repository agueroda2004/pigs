package tests

import (
	"context"

	"github.com/google/uuid"

	operatordomain "server/internal/modules/operator/domain"
)

type fakeOperatorRepository struct {
	exists             bool
	existsErr          error
	getOperator        *operatordomain.Operator
	getErr             error
	listOperators      []operatordomain.OperatorListItem
	listErr            error
	listDropdown       []operatordomain.OperatorDropdown
	listDropdownActive *bool
	listDropdownErr    error
	createErr          error
	updateErr          error
	deleteErr          error
	deleteCalled       bool
	deletedID          uuid.UUID
	created            *operatordomain.Operator
	updated            *operatordomain.Operator
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

func (f *fakeOperatorRepository) List(_ context.Context) ([]operatordomain.OperatorListItem, error) {
	return f.listOperators, f.listErr
}

func (f *fakeOperatorRepository) ListDropdown(_ context.Context, active *bool) ([]operatordomain.OperatorDropdown, error) {
	f.listDropdownActive = active
	return f.listDropdown, f.listDropdownErr
}

func (f *fakeOperatorRepository) Update(_ context.Context, operator *operatordomain.Operator) error {
	f.updated = operator
	return f.updateErr
}

func (f *fakeOperatorRepository) Delete(_ context.Context, id uuid.UUID) error {
	f.deleteCalled = true
	f.deletedID = id
	return f.deleteErr
}
