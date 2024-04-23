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
	repo.On("FindOneById").Return(&user.User{
		Id:    1,
		Email: "hieu@gmail.com",
	}, nil)

	client, closer := serverStubs(ctx, repo)
	defer closer()

	type expectation struct {
		out *user.User
		err error
	}

	tests := map[string]struct {
		in       *user.GetUserByIdRequest
		expected expectation
	}{
		"Invalid_Id": {
			in: &user.GetUserByIdRequest{
				Id: -1,
			},
			expected: expectation{
				out: &user.User{},
				err: status.Error(codes.InvalidArgument, "Id cant be nagative"),
			},
		},
		"Must_Pass": {
			in: &user.GetUserByIdRequest{
				Id: 1,
			},
			expected: expectation{
				out: &user.User{
					Id:    1,
					Email: "hieu@gmail.com",
				},
				err: nil,
			},
		},
	}

	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			ct := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin"))
			out, err := client.GetUser(ct, tt.in)
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
		out *user.User
		err error
	}

	tests := map[string]struct {
		in       *user.GetUserByEmailRequest
		expected expectation
	}{
		"Invalid_Email": {
			in: &user.GetUserByEmailRequest{
				Email: "dsajdhiusadisa",
			},
			expected: expectation{
				out: &user.User{},
				err: status.Error(codes.InvalidArgument, "mail: missing '@' or angle-addr"),
			},
		},
		"Must_Pass": {
			in: &user.GetUserByEmailRequest{
				Email: "hieu@gmail.com",
			},
			expected: expectation{
				out: &user.User{
					Email: "hieu@gmail.com",
				},
				err: nil,
			},
		},
	}

	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			ct := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin"))
			out, err := client.GetUserByEmail(ct, tt.in)
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

	var nagativeNumber int64 = -1
	var invalidTypeSort server.TypeSort = 4

	tests := map[string]struct {
		in       *user.ListUsersRequest
		expected expectation
	}{
		"Must_Pass": {
			in: &user.ListUsersRequest{},
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
		"Invalid_Limit": {
			in: &user.ListUsersRequest{
				Pagination: &server.Pagination{
					Limit: &nagativeNumber,
				},
			},
			expected: expectation{
				out: &user.ListUsersResponse{},
				err: status.Error(codes.InvalidArgument, "invalid ListUsersRequest.Pagination: embedded message failed validation | caused by: invalid Pagination.Limit: value must be greater than or equal to 1"),
			},
		},
		"Invalid_PageSize": {
			in: &user.ListUsersRequest{
				Pagination: &server.Pagination{
					PageSize: &nagativeNumber,
				},
			},
			expected: expectation{
				out: &user.ListUsersResponse{},
				err: status.Error(codes.InvalidArgument, "invalid ListUsersRequest.Pagination: embedded message failed validation | caused by: invalid Pagination.PageSize: value must be greater than or equal to 1"),
			},
		},
		"Invalid_Page": {
			in: &user.ListUsersRequest{
				Pagination: &server.Pagination{
					Page: &nagativeNumber,
				},
			},
			expected: expectation{
				out: &user.ListUsersResponse{},
				err: status.Error(codes.InvalidArgument, "invalid ListUsersRequest.Pagination: embedded message failed validation | caused by: invalid Pagination.Page: value must be greater than or equal to 1"),
			},
		},
		"Invalid_Sort": {
			in: &user.ListUsersRequest{
				Pagination: &server.Pagination{
					Sort: &invalidTypeSort,
				},
			},
			expected: expectation{
				out: &user.ListUsersResponse{},
				err: status.Error(codes.InvalidArgument, "invalid ListUsersRequest.Pagination: embedded message failed validation | caused by: invalid Pagination.Sort: value must be one of the defined enum values"),
			},
		},
	}

	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			ct := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin"))
			out, err := client.ListUsers(ct, tt.in)
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
		out *user.User
		err error
	}

	tests := map[string]struct {
		in       *user.CreateUserRequest
		expected expectation
	}{
		"Exists_Email": {
			in: &user.CreateUserRequest{
				Email:    "hieu@gmail.com",
				Password: "dsadsads",
			},
			expected: expectation{
				out: &user.User{},
				err: status.Error(
					codes.AlreadyExists,
					fmt.Sprintf("Already exisit email:%v", "hieu@gmail.com"),
				),
			},
		},
	}

	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			ct := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin"))
			out, err := client.CreateUser(ct, tt.in)
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
