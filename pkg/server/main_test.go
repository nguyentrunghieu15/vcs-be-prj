package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/auth"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/logger"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ServerRepositoryStubs struct {
	mock.Mock
}

func (c *ServerRepositoryStubs) CheckServerExists(params map[string]interface{}) bool {
	args := c.Called()
	return args.Bool(0)
}

func (c *ServerRepositoryStubs) CountServers(query *string, filter *pb.FilterServer) (int64, error) {
	args := c.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (c *ServerRepositoryStubs) CreateBacth(userId uint64, data []map[string]interface{}) (*pb.ImportServerResponse, error) {
	args := c.Called()
	return args.Get(0).(*pb.ImportServerResponse), args.Error(1)
}

func (c *ServerRepositoryStubs) CreateServer(params map[string]interface{}) (*model.Server, error) {
	args := c.Called()
	return args.Get(0).(*model.Server), args.Error(1)
}

func (c *ServerRepositoryStubs) DeleteOneById(id uuid.UUID) error {
	args := c.Called()
	return args.Error(0)
}

func (c *ServerRepositoryStubs) DeleteOneByName(name string) error {
	args := c.Called()
	return args.Error(0)
}

func (c *ServerRepositoryStubs) FindOneById(id uuid.UUID) (*model.Server, error) {
	args := c.Called()
	return args.Get(0).(*model.Server), args.Error(1)
}

func (c *ServerRepositoryStubs) FindOneByName(name string) (*model.Server, error) {
	args := c.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Server), args.Error(1)
}

func (c *ServerRepositoryStubs) FindServers(req *pb.ListServerRequest) ([]model.Server, error) {
	args := c.Called()
	return args.Get(0).([]model.Server), args.Error(1)
}

func (c *ServerRepositoryStubs) UpdateOneById(id uuid.UUID, params map[string]interface{}) (*model.Server, error) {
	args := c.Called()
	return args.Get(0).(*model.Server), args.Error(1)
}

func (c *ServerRepositoryStubs) UpdateOneByName(name string, params map[string]interface{}) (*model.Server, error) {
	args := c.Called()
	return args.Get(0).(*model.Server), args.Error(1)
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

type ProducerClientStubs struct {
	mock.Mock
}

func (p *ProducerClientStubs) WriteMessage(ctx context.Context, msg ...kafka.Message) error {
	args := p.Called()
	return args.Error(0)
}

func NewServerServiceStubs(repo IServerRepo, kafka ProducerClientInterface) *ServerService {
	return &ServerService{
		ServerRepo: repo,
		l:          &LoggerDecoratorStubs{},
		auhthorize: &auth.Authorizer{},
		validator:  NewServerServiceValidator(),
		kafka:      kafka,
	}
}

func serverStubs(ctx context.Context, repo IServerRepo, kafka ProducerClientInterface) (pb.ServerServiceClient, func()) {
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	pb.RegisterServerServiceServer(baseServer, NewServerServiceStubs(repo, kafka))
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

	client := pb.NewServerServiceClient(conn)

	return client, closer
}

func TestServerService_CreateServer(t *testing.T) {
	ctx := context.Background()
	type expectation struct {
		out *pb.Server
		err error
	}

	type in struct {
		ctx context.Context
		req *pb.CreateServerRequest
	}

	tests := map[string]struct {
		in
		expected expectation
		m        func() *ServerRepositoryStubs
	}{
		"Missing_header": {
			in: in{
				ctx: context.Background(),
				req: &pb.CreateServerRequest{
					Name:   "test",
					Status: pb.CreateServerRequest_ON,
					Ipv4:   "0.0.0.0",
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.Internal, "Can't get role from request"),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				return repo
			},
		},
		"UnAuthorize": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "user")),
				req: &pb.CreateServerRequest{
					Name:   "test",
					Status: pb.CreateServerRequest_ON,
					Ipv4:   "0.0.0.0",
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.PermissionDenied, "Can't create user"),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				return repo
			},
		},
		"Invalid_Input": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.CreateServerRequest{
					Name:   "test",
					Status: pb.CreateServerRequest_ON,
					Ipv4:   "0.0.0sss.0",
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.InvalidArgument, "Field validation for 'Ipv4' failed on the 'ipv4' tag"),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				return repo
			},
		},
		"Already_Server": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.CreateServerRequest{
					Name:   "test",
					Status: pb.CreateServerRequest_ON,
					Ipv4:   "0.0.0.0",
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.AlreadyExists,
					fmt.Sprintf("Already Exists server name: %v", "test")),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				repo.On("FindOneByName").Return(&model.Server{
					ID:     uuid.MustParse("96191014-0f10-4862-b37b-87b0943d2b04"),
					Name:   "test",
					Ipv4:   "0.0.0.0",
					Status: model.On,
				}, nil)
				return repo
			},
		},
		"Must_Pass": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.CreateServerRequest{
					Name:   "test",
					Status: pb.CreateServerRequest_ON,
					Ipv4:   "0.0.0.0",
				},
			},
			expected: expectation{
				out: &pb.Server{
					Id:     "96191014-0f10-4862-b37b-87b0943d2b04",
					Name:   "test",
					Ipv4:   "0.0.0.0",
					Status: pb.Server_ON,
				},
				err: nil,
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}

				repo.On("FindOneByName").Return(nil, nil)
				repo.On("CreateServer").Return(&model.Server{
					ID:     uuid.MustParse("96191014-0f10-4862-b37b-87b0943d2b04"),
					Name:   "test",
					Ipv4:   "0.0.0.0",
					Status: model.On,
				}, nil)
				return repo
			},
		},
	}

	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			client, closer := serverStubs(ctx, tt.m(), nil)
			defer closer()
			got, err := client.CreateServer(tt.in.ctx, tt.in.req)
			if err != nil {
				if tt.expected.err.Error() != err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.err, err)
				}
			} else {
				gotJson, _ := json.Marshal(got)
				nWantJson, _ := json.Marshal(tt.expected.out)
				if !reflect.DeepEqual(gotJson, nWantJson) {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.out, got)
				}
			}
		})
	}
}

func TestServerService_DeleteServerById(t *testing.T) {
	ctx := context.Background()
	type expectation struct {
		out *emptypb.Empty
		err error
	}

	type in struct {
		ctx context.Context
		req *pb.DeleteServerByIdRequest
	}
	var nilServer *model.Server

	tests := map[string]struct {
		in
		expected expectation
		m        func() *ServerRepositoryStubs
	}{
		"Missing_header": {
			in: in{
				ctx: context.Background(),
				req: &pb.DeleteServerByIdRequest{
					Id: "96191014-0f10-4862-b37b-87b0943d2b04",
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.Internal, "Can't get role from request"),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				return repo
			},
		},
		"UnAuthorize": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "user")),
				req: &pb.DeleteServerByIdRequest{
					Id: "96191014-0f10-4862-b37b-87b0943d2b04",
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.PermissionDenied, "Can't delete server"),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				return repo
			},
		},
		"Invalid_request": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.DeleteServerByIdRequest{
					Id: "96191014-0f10--b37b-87b0943d2b04",
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.InvalidArgument, "Field validation for 'Id' failed on the 'uuid4' tag"),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				return repo
			},
		},
		"Not_Found": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.DeleteServerByIdRequest{
					Id: "96191014-0f10-4862-b37b-87b0943d2b04",
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.NotFound, "Not found server"),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				repo.On("FindOneById").Return(nilServer, nil)
				return repo
			},
		},
		"Delete_Err": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.DeleteServerByIdRequest{
					Id: "96191014-0f10-4862-b37b-87b0943d2b04",
				},
			},
			expected: expectation{
				err: status.Error(codes.Internal, fmt.Errorf("").Error()),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				repo.On("FindOneById").Return(&model.Server{}, nil)
				repo.On("DeleteOneById").Return(fmt.Errorf(""))
				return repo
			},
		},
		"Must_Pass": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.DeleteServerByIdRequest{
					Id: "96191014-0f10-4862-b37b-87b0943d2b04",
				},
			},
			expected: expectation{
				err: nil,
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				repo.On("FindOneById").Return(&model.Server{}, nil)
				repo.On("DeleteOneById").Return(nil)
				return repo
			},
		},
	}
	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			client, closer := serverStubs(ctx, tt.m(), nil)
			defer closer()
			_, err := client.DeleteServerById(tt.in.ctx, tt.in.req)
			if err != nil {
				if tt.expected.err.Error() != err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.err, err)
				}
			}
		})
	}
}

func TestServerService_DeleteServerByName(t *testing.T) {
	ctx := context.Background()
	type expectation struct {
		out *emptypb.Empty
		err error
	}

	type in struct {
		ctx context.Context
		req *pb.ExportServerRequest
	}

	tests := map[string]struct {
		in
		expected expectation
		m        func() (*ServerRepositoryStubs, *ProducerClientStubs)
	}{
		"Missing_header": {
			in: in{
				ctx: context.Background(),
				req: &pb.ExportServerRequest{
					UserId: 1,
					File:   &pb.FileExport{FileName: "Test.xlsx"},
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.Internal, "Can't get role from request"),
			},
			m: func() (*ServerRepositoryStubs, *ProducerClientStubs) {
				repo := &ServerRepositoryStubs{}
				return repo, nil
			},
		},
		"UnAuthorize": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "user")),
				req: &pb.ExportServerRequest{
					UserId: 1,
					File:   &pb.FileExport{FileName: "Test.xlsx"},
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.PermissionDenied, "Can't export server"),
			},
			m: func() (*ServerRepositoryStubs, *ProducerClientStubs) {
				repo := &ServerRepositoryStubs{}
				return repo, nil
			},
		},
		"Invalid_request": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.ExportServerRequest{
					UserId: 1,
					File:   &pb.FileExport{},
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.InvalidArgument, "Field validation for 'FileName' failed on the 'required' tag"),
			},
			m: func() (*ServerRepositoryStubs, *ProducerClientStubs) {
				repo := &ServerRepositoryStubs{}
				return repo, nil
			},
		},

		"Must_Pass": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.ExportServerRequest{
					UserId: 1,
					File:   &pb.FileExport{FileName: "Test.xlsx"},
				},
			},
			expected: expectation{
				err: nil,
			},
			m: func() (*ServerRepositoryStubs, *ProducerClientStubs) {
				repo := &ServerRepositoryStubs{}
				repo.On("FindOneByName").Return(&model.Server{}, nil)
				repo.On("DeleteOneByName").Return(nil)

				k := &ProducerClientStubs{}
				k.On("WriteMessage").Return(nil)
				return repo, k
			},
		},
		"Write_Error": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.ExportServerRequest{
					UserId: 1,
					File:   &pb.FileExport{FileName: "Test.xlsx"},
				},
			},
			expected: expectation{
				err: status.Error(codes.Aborted, "Can't export"),
			},
			m: func() (*ServerRepositoryStubs, *ProducerClientStubs) {
				repo := &ServerRepositoryStubs{}
				repo.On("FindOneByName").Return(&model.Server{}, nil)
				repo.On("DeleteOneByName").Return(nil)

				k := &ProducerClientStubs{}
				k.On("WriteMessage").Return(fmt.Errorf(""))
				return repo, k
			},
		},
	}
	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			repo, k := tt.m()
			client, closer := serverStubs(ctx, repo, k)
			defer closer()
			_, err := client.ExportServer(tt.in.ctx, tt.in.req)
			if err != nil {
				if tt.expected.err.Error() != err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.err, err)
				}
			}
		})
	}
}

func TestServerService_GetServerById(t *testing.T) {
	ctx := context.Background()
	type expectation struct {
		out *pb.Server
		err error
	}

	type in struct {
		ctx context.Context
		req *pb.GetServerByIdRequest
	}
	var nilServer *model.Server

	tests := map[string]struct {
		in
		expected expectation
		m        func() *ServerRepositoryStubs
	}{
		"Invalid_request": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.GetServerByIdRequest{
					Id: "96191014-0f10--b37b-87b0943d2b04",
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.InvalidArgument, "Field validation for 'Id' failed on the 'uuid4' tag"),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				return repo
			},
		},
		"Not_Found": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.GetServerByIdRequest{
					Id: "96191014-0f10-4862-b37b-87b0943d2b04",
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.NotFound, "Not found server"),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				repo.On("FindOneById").Return(nilServer, nil)
				return repo
			},
		},
		"Must_Pass": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.GetServerByIdRequest{
					Id: "96191014-0f10-4862-b37b-87b0943d2b04",
				},
			},
			expected: expectation{
				err: nil,
				out: &pb.Server{
					Id:   "96191014-0f10-4862-b37b-87b0943d2b04",
					Name: "Test",
					Ipv4: "0.0.0.0",
				},
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				repo.On("FindOneById").Return(&model.Server{
					ID:   uuid.MustParse("96191014-0f10-4862-b37b-87b0943d2b04"),
					Name: "Test",
					Ipv4: "0.0.0.0",
				}, nil)
				return repo
			},
		},
	}
	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			client, closer := serverStubs(ctx, tt.m(), nil)
			defer closer()
			_, err := client.GetServerById(tt.in.ctx, tt.in.req)
			if err != nil {
				if tt.expected.err.Error() != err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.err, err)
				}
			}
		})
	}
}

func TestServerService_GetServerByName(t *testing.T) {
	ctx := context.Background()
	type expectation struct {
		out *pb.Server
		err error
	}

	type in struct {
		ctx context.Context
		req *pb.GetServerByNameRequest
	}
	var nilServer *model.Server

	tests := map[string]struct {
		in
		expected expectation
		m        func() *ServerRepositoryStubs
	}{
		"Invalid_request": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.GetServerByNameRequest{},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.InvalidArgument, "Field validation for 'Name' failed on the 'required' tag"),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				return repo
			},
		},
		"Not_Found": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.GetServerByNameRequest{
					Name: "Test",
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.NotFound, "Not found server"),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				repo.On("FindOneByName").Return(nilServer, nil)
				return repo
			},
		},
		"Must_Pass": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.GetServerByNameRequest{
					Name: "Test",
				},
			},
			expected: expectation{
				err: nil,
				out: &pb.Server{
					Id:   "96191014-0f10-4862-b37b-87b0943d2b04",
					Name: "Test",
					Ipv4: "0.0.0.0",
				},
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				repo.On("FindOneByName").Return(&model.Server{
					ID:   uuid.MustParse("96191014-0f10-4862-b37b-87b0943d2b04"),
					Name: "Test",
					Ipv4: "0.0.0.0",
				}, nil)
				return repo
			},
		},
	}
	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			client, closer := serverStubs(ctx, tt.m(), nil)
			defer closer()
			_, err := client.GetServerByName(tt.in.ctx, tt.in.req)
			if err != nil {
				if tt.expected.err.Error() != err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.err, err)
				}
			}
		})
	}
}

func TestServerService_ListServers(t *testing.T) {
	ctx := context.Background()
	type expectation struct {
		out *pb.ListServersResponse
		err error
	}

	type in struct {
		ctx context.Context
		req *pb.ListServerRequest
	}
	var negativeNum int64 = -1

	tests := map[string]struct {
		in
		expected expectation
		m        func() *ServerRepositoryStubs
	}{
		"Invalid_request": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.ListServerRequest{
					Pagination: &pb.Pagination{
						Limit: &negativeNum,
					},
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.InvalidArgument, "Field validation for 'Limit' failed on the 'min' tag"),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				return repo
			},
		},
		"Find_Errs": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.ListServerRequest{},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.Internal, fmt.Errorf("").Error()),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				repo.On("FindServers").Return([]model.Server{}, fmt.Errorf(""))
				return repo
			},
		},
		"Must_Pass": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.ListServerRequest{},
			},
			expected: expectation{
				err: nil,
				out: &pb.ListServersResponse{
					Servers: []*pb.Server{
						{
							Id:   "96191014-0f10-4862-b37b-87b0943d2b04",
							Name: "Test",
							Ipv4: "0.0.0.0",
						},
					},
					Total: 1,
				},
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				repo.On("FindServers").Return([]model.Server{
					{
						ID:   uuid.MustParse("96191014-0f10-4862-b37b-87b0943d2b04"),
						Name: "Test",
						Ipv4: "0.0.0.0",
					},
				}, nil)
				repo.On("CountServers").Return(int64(1), nil)
				return repo
			},
		},
	}
	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			client, closer := serverStubs(ctx, tt.m(), nil)
			defer closer()
			_, err := client.ListServers(tt.in.ctx, tt.in.req)
			if err != nil {
				if tt.expected.err.Error() != err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.err, err)
				}
			}
		})
	}
}

func TestServerService_UpdateServer(t *testing.T) {
	ctx := context.Background()
	type expectation struct {
		out *pb.Server
		err error
	}

	type in struct {
		ctx context.Context
		req *pb.UpdateServerRequest
	}

	var nilServer *model.Server

	tests := map[string]struct {
		in
		expected expectation
		m        func() *ServerRepositoryStubs
	}{
		"Missing_header": {
			in: in{
				ctx: context.Background(),
				req: &pb.UpdateServerRequest{
					Id:     "96191014-0f10-4862-b37b-87b0943d2b04",
					Name:   "test",
					Status: pb.UpdateServerRequest_ON,
					Ipv4:   "0.0.0.0",
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.Internal, "Can't get role from request"),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				return repo
			},
		},
		"UnAuthorize": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "user")),
				req: &pb.UpdateServerRequest{
					Id:     "96191014-0f10-4862-b37b-87b0943d2b04",
					Name:   "test",
					Status: pb.UpdateServerRequest_ON,
					Ipv4:   "0.0.0.0",
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.PermissionDenied, "Can't update user"),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				return repo
			},
		},
		"Invalid_Input": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.UpdateServerRequest{
					Id:     "96191014-0f10-4862-b37b-87b0943d2b04",
					Name:   "test",
					Status: pb.UpdateServerRequest_ON,
					Ipv4:   "0.0.0.s0",
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.InvalidArgument, "Field validation for 'Ipv4' failed on the 'ipv4' tag"),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				return repo
			},
		},
		"Not_Found_Server": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.UpdateServerRequest{
					Id:     "96191014-0f10-4862-b37b-87b0943d2b04",
					Name:   "test",
					Status: pb.UpdateServerRequest_ON,
					Ipv4:   "0.0.0.0",
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.AlreadyExists, fmt.Errorf("Not found server name :%v", "test").Error()),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}
				repo.On("FindOneByName").Return(nilServer, nil)
				return repo
			},
		},
		"Must_Pass": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.UpdateServerRequest{
					Id:     "96191014-0f10-4862-b37b-87b0943d2b04",
					Name:   "test",
					Status: pb.UpdateServerRequest_ON,
					Ipv4:   "0.0.0.0",
				},
			},
			expected: expectation{
				out: &pb.Server{
					Id:     "96191014-0f10-4862-b37b-87b0943d2b04",
					Name:   "test",
					Ipv4:   "0.0.0.0",
					Status: pb.Server_ON,
				},
				err: nil,
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}

				repo.On("FindOneByName").Return(&model.Server{}, nil)
				repo.On("UpdateOneById").Return(&model.Server{
					ID:     uuid.MustParse("96191014-0f10-4862-b37b-87b0943d2b04"),
					Name:   "test",
					Ipv4:   "0.0.0.0",
					Status: model.On,
				}, nil)
				return repo
			},
		},
		"Err_Update": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.UpdateServerRequest{
					Id:     "96191014-0f10-4862-b37b-87b0943d2b04",
					Name:   "test",
					Status: pb.UpdateServerRequest_ON,
					Ipv4:   "0.0.0.0",
				},
			},
			expected: expectation{
				out: nil,
				err: status.Error(codes.Internal, fmt.Errorf("").Error()),
			},
			m: func() *ServerRepositoryStubs {
				repo := &ServerRepositoryStubs{}

				repo.On("FindOneByName").Return(&model.Server{}, nil)
				repo.On("UpdateOneById").Return(nilServer, fmt.Errorf(""))
				return repo
			},
		},
	}

	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			client, closer := serverStubs(ctx, tt.m(), nil)
			defer closer()
			got, err := client.UpdateServer(tt.in.ctx, tt.in.req)
			if err != nil {
				fmt.Println(err)
				if tt.expected.err.Error() != err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.err, err)
				}
			} else {
				gotJson, _ := json.Marshal(got)
				nWantJson, _ := json.Marshal(tt.expected.out)
				if !reflect.DeepEqual(gotJson, nWantJson) {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.out, got)
				}
			}
		})
	}
}
