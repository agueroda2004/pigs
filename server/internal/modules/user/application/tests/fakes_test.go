package tests

import (
	"context"

	"github.com/google/uuid"

	userdomain "server/internal/modules/user/domain"
)

type fakeUserRepository struct {
	exists    bool
	existsErr error
	getUser   *userdomain.User
	getErr    error
	createErr error
	updateErr error
	created   *userdomain.User
	updated   *userdomain.User
}

func (f *fakeUserRepository) Create(_ context.Context, user *userdomain.User) error {
	f.created = user
	return f.createErr
}

func (f *fakeUserRepository) GetByID(_ context.Context, _ uuid.UUID) (*userdomain.User, error) {
	return f.getUser, f.getErr
}

func (f *fakeUserRepository) ExistsByUsername(_ context.Context, _ string) (bool, error) {
	return f.exists, f.existsErr
}

func (f *fakeUserRepository) Update(_ context.Context, user *userdomain.User) error {
	f.updated = user
	return f.updateErr
}

type fakePasswordHasher struct {
	hash     string
	err      error
	password string
}

func (f *fakePasswordHasher) Hash(password string) (string, error) {
	f.password = password
	return f.hash, f.err
}
