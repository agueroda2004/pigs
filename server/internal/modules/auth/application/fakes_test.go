package application

import (
	"context"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	userdomain "server/internal/modules/user/domain"
)

type fakeUserReader struct {
	findUser *userdomain.User
	findErr  error
	getUser  *userdomain.User
	getErr   error
}

func (f *fakeUserReader) FindByUsername(_ context.Context, _ string) (*userdomain.User, error) {
	return f.findUser, f.findErr
}

func (f *fakeUserReader) GetByID(_ context.Context, _ uuid.UUID) (*userdomain.User, error) {
	return f.getUser, f.getErr
}

type fakePasswordVerifier struct {
	err      error
	hashed   string
	password string
}

func (f *fakePasswordVerifier) Verify(hashedPassword string, password string) error {
	f.hashed = hashedPassword
	f.password = password
	return f.err
}

type fakeAccessTokenIssuer struct {
	token string
	err   error
	user  authdomain.AuthenticatedUser
}

func (f *fakeAccessTokenIssuer) Issue(user authdomain.AuthenticatedUser) (string, error) {
	f.user = user
	return f.token, f.err
}

type fakeTokenGenerator struct {
	token string
	err   error
}

func (f *fakeTokenGenerator) Generate() (string, error) {
	return f.token, f.err
}

type fakeTokenHasher struct {
	hash   string
	err    error
	tokens []string
}

func (f *fakeTokenHasher) Hash(token string) (string, error) {
	f.tokens = append(f.tokens, token)
	return f.hash, f.err
}

type fakeRefreshTokenRepository struct {
	created       []*authdomain.RefreshToken
	get           *authdomain.RefreshToken
	getErr        error
	usedID        uuid.UUID
	useErr        error
	revokedFamily uuid.UUID
	revokeErr     error
	revokedByHash string
}

func (f *fakeRefreshTokenRepository) Create(_ context.Context, token *authdomain.RefreshToken) error {
	f.created = append(f.created, token)
	return nil
}

func (f *fakeRefreshTokenRepository) GetByHash(_ context.Context, _ string) (*authdomain.RefreshToken, error) {
	return f.get, f.getErr
}

func (f *fakeRefreshTokenRepository) MarkAsUsed(_ context.Context, id uuid.UUID) error {
	f.usedID = id
	return f.useErr
}

func (f *fakeRefreshTokenRepository) RevokeFamily(_ context.Context, familyID uuid.UUID) error {
	f.revokedFamily = familyID
	return f.revokeErr
}

func (f *fakeRefreshTokenRepository) RevokeByHash(_ context.Context, hash string) error {
	f.revokedByHash = hash
	return nil
}
