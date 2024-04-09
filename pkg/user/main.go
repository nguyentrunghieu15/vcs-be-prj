package user

import (
	"context"

	"github.com/nguyentrunghieu15/vcs-common-prj/apu/user"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserServer struct {
	user.UserServiceServer
}

func (u *UserServer) ListUsers(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	return nil, nil
}
func (u *UserServer) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.User, error) {
	return nil, nil
}
func (u *UserServer) UpdateUser(ctx context.Context, req *user.UpdateUserByIdRequest) (*user.User, error) {
	return nil, nil
}
func (u *UserServer) DeleteUSer(ctx context.Context, req *user.DeleteUserByIdRequest) (*emptypb.Empty, error) {
	return nil, nil
}
