package tests

import (
	"context"

	"github.com/google/uuid"

	breeddomain "server/internal/modules/breed/domain"
	"server/internal/modules/breed/ports"
)

type fakeBreedRepository struct {
	exists              bool
	existsErr           error
	existsExcluding     bool
	existsExcludingErr  error
	existsExcludingName string
	existsExcludingID   uuid.UUID
	getBreed            *breeddomain.Breed
	getErr              error
	listBreeds          []*breeddomain.Breed
	listErr             error
	listFilter          ports.BreedFilter
	listDropdown        []breeddomain.BreedDropdown
	listDropdownErr     error
	listDropdownActive  bool
	createErr           error
	updateErr           error
	deleteErr           error
	deletedID           uuid.UUID
	created             *breeddomain.Breed
	updated             *breeddomain.Breed
}

func (f *fakeBreedRepository) Create(_ context.Context, breed *breeddomain.Breed) error {
	f.created = breed
	return f.createErr
}

func (f *fakeBreedRepository) GetByID(_ context.Context, _ uuid.UUID) (*breeddomain.Breed, error) {
	return f.getBreed, f.getErr
}

func (f *fakeBreedRepository) ExistsByName(_ context.Context, _ string) (bool, error) {
	return f.exists, f.existsErr
}

func (f *fakeBreedRepository) ExistsByNameExcludingID(_ context.Context, name string, id uuid.UUID) (bool, error) {
	f.existsExcludingName = name
	f.existsExcludingID = id
	return f.existsExcluding, f.existsExcludingErr
}

func (f *fakeBreedRepository) List(_ context.Context, filter ports.BreedFilter) ([]*breeddomain.Breed, error) {
	f.listFilter = filter
	return f.listBreeds, f.listErr
}

func (f *fakeBreedRepository) ListDropdown(_ context.Context, active bool) ([]breeddomain.BreedDropdown, error) {
	f.listDropdownActive = active
	return f.listDropdown, f.listDropdownErr
}

func (f *fakeBreedRepository) Update(_ context.Context, breed *breeddomain.Breed) error {
	f.updated = breed
	return f.updateErr
}

func (f *fakeBreedRepository) Delete(_ context.Context, id uuid.UUID) error {
	f.deletedID = id
	return f.deleteErr
}
