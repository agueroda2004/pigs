package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	userapplication "server/internal/modules/user/application"
	userdomain "server/internal/modules/user/domain"
	userinfra "server/internal/modules/user/infrastructure"
)

type fakeCreateUserUseCase struct {
	user    *userdomain.User
	err     error
	command userapplication.CreateUserCommand
	called  bool
}

func (f *fakeCreateUserUseCase) Execute(_ context.Context, command userapplication.CreateUserCommand) (*userdomain.User, error) {
	f.called = true
	f.command = command
	return f.user, f.err
}

type fakeListUsersUseCase struct {
	users  []*userdomain.User
	err    error
	called bool
}

func (f *fakeListUsersUseCase) Execute(_ context.Context) ([]*userdomain.User, error) {
	f.called = true
	return f.users, f.err
}

type fakeUpdateUserUseCase struct {
	user    *userdomain.User
	err     error
	userID  uuid.UUID
	command userapplication.UpdateOwnUserCommand
	called  bool
}

func (f *fakeUpdateUserUseCase) Execute(_ context.Context, userID uuid.UUID, command userapplication.UpdateOwnUserCommand) (*userdomain.User, error) {
	f.called = true
	f.userID = userID
	f.command = command
	return f.user, f.err
}

type fakeUpdateUserByAdminUseCase struct {
	user    *userdomain.User
	err     error
	userID  uuid.UUID
	command userapplication.UpdateUserByAdminCommand
	called  bool
}

func (f *fakeUpdateUserByAdminUseCase) Execute(_ context.Context, userID uuid.UUID, command userapplication.UpdateUserByAdminCommand) (*userdomain.User, error) {
	f.called = true
	f.userID = userID
	f.command = command
	return f.user, f.err
}

func newTestHandler(create userinfra.CreateUserUseCase, update userinfra.UpdateOwnUserUseCase, updateAdmin userinfra.UpdateUserByAdminUseCase) *userinfra.UserHandler {
	return newTestHandlerWithList(&fakeListUsersUseCase{}, create, update, updateAdmin)
}

func newTestHandlerWithList(list userinfra.ListUsersUseCase, create userinfra.CreateUserUseCase, update userinfra.UpdateOwnUserUseCase, updateAdmin userinfra.UpdateUserByAdminUseCase) *userinfra.UserHandler {
	return userinfra.NewUserHandler(create, list, update, updateAdmin, func(next http.Handler) http.Handler { return next })
}

func authenticatedRequest(request *http.Request, userID uuid.UUID) *http.Request {
	return request.WithContext(authdomain.WithAuthenticatedUser(request.Context(), authdomain.AuthenticatedUser{
		UserID: userID,
		Role:   userdomain.RoleAdmin,
	}))
}

func serve(handler *userinfra.UserHandler, request *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func handlerUser() *userdomain.User {
	return &userdomain.User{
		ID: uuid.New(), Name: "Ana", Username: "ana", Role: userdomain.RoleUser, Active: true,
		CreatedAt: time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC),
		UpdatedAt: time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC),
	}
}
