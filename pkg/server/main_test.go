package server

import (
	"context"
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
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
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

func NewServerServiceStubs(repo IServerRepo) *ServerService {
	return &ServerService{
		ServerRepo: repo,
		l:          &LoggerDecoratorStubs{},
		auhthorize: &auth.Authorizer{},
		validator:  NewServerServiceValidator(),
		kafka:      nil,
	}
}

func serverStubs(ctx context.Context, repo IServerRepo) (pb.ServerServiceClient, func()) {
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	pb.RegisterServerServiceServer(baseServer, NewServerServiceStubs(repo))
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
	repo := &ServerRepositoryStubs{}
	var nilServerModel model.Server
	repo.On("CreateServer").Return(&model.Server{
		ID:     uuid.MustParse("96191014-0f10-4862-b37b-87b0943d2b04"),
		Name:   "test",
		Ipv4:   "0.0.0.0",
		Status: model.On,
	}, nil)

	client, closer := serverStubs(ctx, repo)
	defer closer()

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
	}{
		"Must_Pass": {
			in: in{
				ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "admin")),
				req: &pb.CreateServerRequest{
					Name:   "test",
					Status: pb.CreateServerRequest_OFF,
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
		},
	}

	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			if scenario == "Not_Found" {
				repo.On("FindOneByName").Return(&nilServerModel, fmt.Errorf("Not Found"))
			} else {
				repo.On("FindOneByName").Return(nil, nil)
			}
			got, err := client.CreateServer(tt.in.ctx, tt.in.req)
			fmt.Println(got)
			if err != nil {
				if tt.expected.err.Error() != err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.err, err)
				}
			} else {
				if !reflect.DeepEqual(got, tt.expected.out) {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.out, got)
				}
			}
		})
	}
}
