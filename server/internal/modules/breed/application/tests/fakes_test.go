package tests

import (
	"context"

	"github.com/google/uuid"

	breeddomain "server/internal/modules/breed/domain"
)

type fakeBreedRepository struct {
	exists         bool
	existsErr      error
	getBreed       *breeddomain.Breed
	getErr         error
	listBreeds     []*breeddomain.Breed
	listErr        error
	listOptions    []breeddomain.BreedOption
	listOptionsErr error
	createErr      error
	updateErr      error
	created        *breeddomain.Breed
	updated        *breeddomain.Breed
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

func (f *fakeBreedRepository) List(_ context.Context) ([]*breeddomain.Breed, error) {
	return f.listBreeds, f.listErr
}

func (f *fakeBreedRepository) ListActiveOptions(_ context.Context) ([]breeddomain.BreedOption, error) {
	return f.listOptions, f.listOptionsErr
}

func (f *fakeBreedRepository) Update(_ context.Context, breed *breeddomain.Breed) error {
	f.updated = breed
	return f.updateErr
}
