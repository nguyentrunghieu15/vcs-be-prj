package user

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/auth"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/logger"
	"github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/apu/user"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserRepositoryStubs struct {
	mock.Mock
}

func (c *UserRepositoryStubs) CreateUser(u map[string]interface{}) (*model.User, error) {
	args := c.Called()
	return args.Get(0).(*model.User), args.Error(1)
}
func (c *UserRepositoryStubs) DeleteOneByEmail(email string) error {
	args := c.Called()
	return args.Error(0)
}
func (c *UserRepositoryStubs) DeleteOneById(id int) error {
	args := c.Called()
	return args.Error(0)
}
func (c *UserRepositoryStubs) FindOneByEmail(email string) (*model.User, error) {
	args := c.Called()
	return args.Get(0).(*model.User), args.Error(1)
}
func (c *UserRepositoryStubs) FindOneById(id int) (*model.User, error) {
	args := c.Called()
	return args.Get(0).(*model.User), args.Error(1)
}
func (c *UserRepositoryStubs) FindUsers(req *user.ListUsersRequest) ([]model.User, error) {
	args := c.Called()
	return args.Get(0).([]model.User), args.Error(1)

}
func (c *UserRepositoryStubs) UpdateOneByEmail(email string, u map[string]interface{}) (*model.User, error) {
	args := c.Called()
	return args.Get(0).(*model.User), args.Error(1)
}
func (c *UserRepositoryStubs) UpdateOneById(id int, u map[string]interface{}) (*model.User, error) {
	args := c.Called()
	return args.Get(0).(*model.User), args.Error(1)
}

type LoggerDecoratorStubs struct {
	mock.Mock
}

func (logger *LoggerDecoratorStubs) ImplementedMiddlewareLogger() echo.MiddlewareFunc {
	args := logger.Called()
	return args.Get(0).(echo.MiddlewareFunc)
}
func (logger *LoggerDecoratorStubs) Log(level logger.LevelLogType, fields map[string]interface{}) {
	return
}
func (logger *LoggerDecoratorStubs) SetOutput(w io.Writer) {
	return
}
func (logger *LoggerDecoratorStubs) SetToday(today time.Time) {
	return
}

func NewUserServerStubs(repo UserRepositoryDecoratorInterface) *UserServer {
	return &UserServer{
		UserRepo:  repo,
		l:         &LoggerDecoratorStubs{},
		bcrypt:    &auth.BcryptService{},
		authorize: &auth.Authorizer{},
		validator: NewUserServiceValidator(),
	}
}

func serverStubs(ctx context.Context, repo UserRepositoryDecoratorInterface) (user.UserServiceClient, func()) {
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	user.RegisterUserServiceServer(baseServer, NewUserServerStubs(repo))
	go func() {
		if err := baseServer.Serve(lis); err != nil {
			log.Printf("error serving server: %v", err)
		}
	}()

	conn, err := grpc.DialContext(ctx, "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("error connecting to server: %v", err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			log.Printf("error closing listener: %v", err)
		}
		baseServer.Stop()
	}

	client := user.NewUserServiceClient(conn)

	return client, closer
}

func TestUserServer_GetUser(t *testing.T) {
	ctx := context.Background()
	repo := &UserRepositoryStubs{}
	repo.On("FindOneById").Return(&model.User{
		Email: "hieu@gmail.com",
		// Roles: model.RoleAdmin,
	}, nil)

	client, closer := serverStubs(ctx, repo)
	defer closer()

	type expectation struct {
		out *user.ResponseUser
		err error
	}

	type in struct {
		ctx context.Context
		req *user.GetUserByIdRequest
	}

	tests := map[string]struct {
		in
		expected expectation
	}{
		"Non_Role": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("test", "test")),
				req: &user.GetUserByIdRequest{
					Id: 1,
				},
			},
			expected: expectation{
				out: &user.ResponseUser{},
				err: status.Error(codes.Internal, "Can't get role from request"),
			},
		},
		"Invalid_Request": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &user.GetUserByIdRequest{
					Id: -1,
				},
			},
			expected: expectation{
				out: &user.ResponseUser{},
				err: status.Error(codes.InvalidArgument, "Id cant be nagative"),
			},
		},
		"Must_Pass": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &user.GetUserByIdRequest{
					Id: 1,
				},
			},
			expected: expectation{
				out: &user.ResponseUser{
					Email: "hieu@gmail.com",
					// Roles: user.ResponseUser_admin,
				},
				err: nil,
			},
		},
	}

	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			out, err := client.GetUser(tt.in.ctx, tt.in.req)
			if err != nil {
				if tt.expected.err.Error() != err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.err, err)
				}
			} else {
				if tt.expected.out.Id != out.Id ||
					tt.expected.out.Email != out.Email {
					t.Errorf("Out -> \nWant: %q\nGot : %q", tt.expected.out, out)
				}
			}

		})
	}
}

func TestUserServer_GetUserByEmail(t *testing.T) {
	ctx := context.Background()
	repo := &UserRepositoryStubs{}
	repo.On("FindOneByEmail").Return(&model.User{
		Email: "hieu@gmail.com",
	}, nil)

	client, closer := serverStubs(ctx, repo)
	defer closer()

	type expectation struct {
		out *user.ResponseUser
		err error
	}

	type in struct {
		ctx context.Context
		req *user.GetUserByEmailRequest
	}

	tests := map[string]struct {
		in
		expected expectation
	}{
		"Non_Role": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("test", "test")),
				req: &user.GetUserByEmailRequest{
					Email: "hieu@gmail.com",
				},
			},
			expected: expectation{
				out: &user.ResponseUser{},
				err: status.Error(codes.Internal, "Can't get role from request"),
			},
		},
		"Invalid_Request": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &user.GetUserByEmailRequest{
					Email: "dsajdhiusadisa",
				},
			},
			expected: expectation{
				out: &user.ResponseUser{},
				err: status.Error(codes.InvalidArgument, "Field validation for 'Email' failed on the 'email' tag"),
			},
		},
		"Must_Pass": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &user.GetUserByEmailRequest{
					Email: "hieu@gmail.com",
				},
			},
			expected: expectation{
				out: &user.ResponseUser{
					Email: "hieu@gmail.com",
				},
				err: nil,
			},
		},
	}

	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			out, err := client.GetUserByEmail(tt.in.ctx, tt.in.req)
			if err != nil {
				if tt.expected.err.Error() != err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.err, err)
				}
			} else {
				if tt.expected.out.Id != out.Id ||
					tt.expected.out.Email != out.Email {
					t.Errorf("Out -> \nWant: %q\nGot : %q", tt.expected.out, out)
				}
			}

		})
	}
}

func TestUserServer_ListUsers(t *testing.T) {
	ctx := context.Background()
	repo := &UserRepositoryStubs{}
	repo.On("FindUsers").Return([]model.User{
		{Email: "hieu@gmail.com"},
	}, nil)

	client, closer := serverStubs(ctx, repo)
	defer closer()

	type expectation struct {
		out *user.ListUsersResponse
		err error
	}

	var nagativeNum int64 = -1
	var positiveNum int64 = 1

	type in struct {
		ctx context.Context
		req *user.ListUsersRequest
	}

	tests := map[string]struct {
		in
		expected expectation
	}{
		"Non_Role": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("test", "test")),
				req: &user.ListUsersRequest{},
			},
			expected: expectation{
				out: &user.ListUsersResponse{},
				err: status.Error(codes.Internal, "Can't get role from request"),
			},
		},
		"Invalid_Request": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &user.ListUsersRequest{
					Pagination: &server.Pagination{
						Limit: &nagativeNum,
					},
				},
			},
			expected: expectation{
				out: &user.ListUsersResponse{},
				err: status.Error(codes.InvalidArgument, "Field validation for 'Limit' failed on the 'min' tag"),
			},
		},
		"Must_Pass": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &user.ListUsersRequest{
					Pagination: &server.Pagination{
						Limit: &positiveNum,
					},
				},
			},
			expected: expectation{
				out: &user.ListUsersResponse{
					Users: []*user.ResponseUser{
						&user.ResponseUser{
							Email: "hieu@gmail.com",
						},
					},
				},
				err: nil,
			},
		},
	}

	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			out, err := client.ListUsers(tt.in.ctx, tt.in.req)
			if err != nil {
				if tt.expected.err.Error() != err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.err, err)
				}
			} else {
				if tt.expected.out.Users[0].Email != out.Users[0].Email {
					t.Errorf("Out -> \nWant: %q\nGot : %q", tt.expected.out, out)
				}
			}

		})
	}
}

func TestUserServer_CreateUser(t *testing.T) {
	ctx := context.Background()
	repo := &UserRepositoryStubs{}
	repo.On("CreateUser").Return(&model.User{
		Email: "hieu@gmail.com",
	}, nil)
	var nilUserModel model.User
	repo.On("FindOneByEmail").Return(&nilUserModel, nil)

	client, closer := serverStubs(ctx, repo)
	defer closer()

	type expectation struct {
		out *user.ResponseUser
		err error
	}

	type in struct {
		ctx context.Context
		req *user.CreateUserRequest
	}

	tests := map[string]struct {
		in
		expected expectation
	}{
		"Non_Role": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("test", "test")),
				req: &user.CreateUserRequest{},
			},
			expected: expectation{
				out: &user.ResponseUser{},
				err: status.Error(codes.Internal, "Can't get role from request"),
			},
		},
		"Invalid_Request": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &user.CreateUserRequest{
					Email:    "dasdasdasdsa",
					Password: "dsadsadasds",
				},
			},
			expected: expectation{
				out: &user.ResponseUser{},
				err: status.Error(codes.InvalidArgument, "Field validation for 'Email' failed on the 'email' tag"),
			},
		},
		"Permisson_Denied": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "user")),
				req: &user.CreateUserRequest{
					Email:    "dasdasdasdsa",
					Password: "dsadsadasds",
				},
			},
			expected: expectation{
				out: &user.ResponseUser{},
				err: status.Error(codes.PermissionDenied, "Can't create user"),
			},
		},
		"Alleady_Email": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &user.CreateUserRequest{
					Email:    "hieu@gmail.com",
					Password: "adaygduwdsad",
				},
			},
			expected: expectation{
				out: &user.ResponseUser{
					Email: "hieu@gmail.com",
				},
				err: status.Error(
					codes.AlreadyExists,
					fmt.Sprintf("Already exisit email:%v", "hieu@gmail.com"),
				),
			},
		},
		"Must_Pass": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &user.CreateUserRequest{
					Email:    "hieu1@gmail.com",
					Password: "adaygduwdsad",
				},
			},
			expected: expectation{
				out: &user.ResponseUser{
					Email: "hieu@gmail.com",
				},
				err: status.Error(
					codes.AlreadyExists,
					fmt.Sprintf("Already exisit email:%v", "hieu@gmail.com"),
				),
			},
		},
	}

	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			out, err := client.CreateUser(tt.in.ctx, tt.in.req)
			if err != nil {
				if tt.expected.err.Error() != err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.err, err)
				}
			} else {
				if tt.expected.out.Email != out.Email {
					t.Errorf("Out -> \nWant: %q\nGot : %q", tt.expected.out, out)
				}
			}

		})
	}
}

func TestUserServer_UpdateUser(t *testing.T) {
	ctx := context.Background()
	repo := &UserRepositoryStubs{}
	repo.On("UpdateOneById").Return(&model.User{
		Email: "hieu@gmail.com",
	}, nil)
	var nilUserModel model.User

	client, closer := serverStubs(ctx, repo)
	defer closer()

	type expectation struct {
		out *user.ResponseUser
		err error
	}

	type in struct {
		ctx context.Context
		req *user.UpdateUserByIdRequest
	}

	tests := map[string]struct {
		in
		expected expectation
	}{
		"Non_Role": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("test", "test")),
				req: &user.UpdateUserByIdRequest{},
			},
			expected: expectation{
				out: &user.ResponseUser{},
				err: status.Error(codes.Internal, "Can't get role from request"),
			},
		},
		"Invalid_Request": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &user.UpdateUserByIdRequest{
					Email: "dasdasdasdsa",
					Id:    1,
				},
			},
			expected: expectation{
				out: &user.ResponseUser{},
				err: status.Error(codes.InvalidArgument, "Field validation for 'Email' failed on the 'email' tag"),
			},
		},
		"Permisson_Denied": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "user")),
				req: &user.UpdateUserByIdRequest{
					Id:    1,
					Email: "dasdasdasdsa",
				},
			},
			expected: expectation{
				out: &user.ResponseUser{},
				err: status.Error(codes.PermissionDenied, "Can't update user"),
			},
		},
		"Not_Found": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &user.UpdateUserByIdRequest{
					Id:    10,
					Email: "hieu@gmail.com",
				},
			},
			expected: expectation{
				out: &user.ResponseUser{
					Email: "hieu@gmail.com",
				},
				err: status.Error(codes.NotFound, fmt.Errorf("Not Found").Error()),
			},
		},
		"Must_Pass": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &user.UpdateUserByIdRequest{
					Id:    1,
					Email: "hieu1@gmail.com",
				},
			},
			expected: expectation{
				out: &user.ResponseUser{
					Email: "hieu@gmail.com",
				},
				err: nil,
			},
		},
	}

	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			if scenario == "Not_Found" {
				repo.On("FindOneById").Return(&nilUserModel, fmt.Errorf("Not Found"))
			} else {
				repo.On("FindOneById").Return(&nilUserModel, nil)
			}
			out, err := client.UpdateUser(tt.in.ctx, tt.in.req)
			if err != nil {
				if tt.expected.err.Error() != err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.err, err)
				}
			} else {
				if tt.expected.out.Email != out.Email {
					t.Errorf("Out -> \nWant: %q\nGot : %q", tt.expected.out, out)
				}
			}

		})
	}
}

func TestUserServer_DeleteUser(t *testing.T) {
	ctx := context.Background()
	repo := &UserRepositoryStubs{}
	var nilUserModel model.User
	repo.On("DeleteOneById").Return(nil)

	client, closer := serverStubs(ctx, repo)
	defer closer()

	type expectation struct {
		out *emptypb.Empty
		err error
	}

	type in struct {
		ctx context.Context
		req *user.DeleteUserByIdRequest
	}

	tests := map[string]struct {
		in
		expected expectation
	}{
		"Non_Role": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("test", "test")),
				req: &user.DeleteUserByIdRequest{
					Id: 1,
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.Internal, "Can't get role from request"),
			},
		},
		"Invalid_Request": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &user.DeleteUserByIdRequest{
					Id: -1,
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.InvalidArgument, "Field validation for 'Id' failed on the 'min' tag"),
			},
		},
		"Permisson_Denied": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "user")),
				req: &user.DeleteUserByIdRequest{
					Id: 1,
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.PermissionDenied, "Can't delete user"),
			},
		},
		"Not_Found": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &user.DeleteUserByIdRequest{
					Id: 10,
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.NotFound, fmt.Errorf("Not Found").Error()),
			},
		},
		"Must_Pass": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &user.DeleteUserByIdRequest{
					Id: 1,
				},
			},
			expected: expectation{
				out: nil,
				err: nil,
			},
		},
	}

	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			if scenario == "Not_Found" {
				repo.On("FindOneById").Return(&nilUserModel, fmt.Errorf("Not Found"))
			} else {
				repo.On("FindOneById").Return(&nilUserModel, nil)
			}
			_, err := client.DeleteUser(tt.in.ctx, tt.in.req)
			if err != nil {
				if tt.expected.err.Error() != err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.err, err)
				}
			}
		})
	}
}
