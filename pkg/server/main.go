package server

import (
	"context"
	"fmt"
	"log"

	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/logger"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/user"
	"github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/managedb"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type ServerService struct {
	pb.ServerServiceServer
	l          *logger.LoggerDecorator
	ServerRepo *ServerRepositoryDecorator
}

func NewServerService() *ServerService {
	dsnPostgres := fmt.Sprintf(
		"host=%v user=%v password=%v dbname=%v port=%v sslmode=%v",
		env.GetEnv("POSTGRES_ADDRESS"),
		env.GetEnv("POSTGRES_USERNAME"),
		env.GetEnv("POSTGRES_PASSWORD"),
		env.GetEnv("POSTGRES_DATABASE"),
		env.GetEnv("POSTGRES_PORT"),
		env.GetEnv("POSTGRES_SSLMODE"),
	)
	postgres, err := managedb.GetConnection(
		managedb.Connection{
			Context: &managedb.PostgreContext{},
			Dsn:     dsnPostgres,
		})
	if err != nil {
		log.Fatalf("AuthService : Can't connect to PostgresSQL Database :%v", err)
	}
	log.Println("Connected database")
	connPostgres, _ := postgres.(*gorm.DB)
	// connPostgres.Config.Logger = gormLogger.Default.LogMode(gormLogger.Silent)
	newLogger := logger.NewLogger()
	newLogger.Config = logger.LoggerConfig{
		IsLogRotate:     true,
		PathToLog:       env.GetEnv("SERVER_LOG_PATH").(string),
		FileNameLogBase: env.GetEnv("SERVER_NAME_FILE_LOG").(string),
	}
	return &ServerService{
		ServerRepo: NewServerRepository(connPostgres),
		l:          newLogger,
	}
}

type LogMessageServer map[string]interface{}

func (s *ServerService) CreateServer(ctx context.Context, req *pb.CreateServerRequest) (*pb.Server, error) {
	s.l.Log(
		logger.INFO,
		LogMessageServer{
			"Action": "Invoked create server",
			"Name":   req.GetName(),
		},
	)

	// validate data
	if err := req.Validate(); err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Create server",
				"Name":   req.GetName(),
				"Error":  "Bad request",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// check already exists server
	if existsServer, _ := s.ServerRepo.FindOneByName(req.GetName()); existsServer != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Create server",
				"Name":   req.GetName(),
				"Error":  "Already Exists",
				"Detail": fmt.Errorf("Already Exists server name", req.GetName()),
			},
		)
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Already Exists server name", req.GetName()))
	}

	// Parse data request
	server, err := ParseMapCreateServerRequest(req)
	if err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Create server",
				"Name":   req.GetName(),
				"Error":  "Parse data",
				"Detail": fmt.Errorf("Parse data", err),
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Get header
	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Create Server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
	}

	if v, ok := header["id"]; ok {
		server["CreatedBy"] = v[0]
	}

	// create server
	createdServer, err := s.ServerRepo.CreateServer(server)
	if err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Create server",
				"Name":   req.GetName(),
				"Error":  "Can't create server",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return ConvertServerModelToServerProto(*createdServer), nil
}
func (s *ServerService) DeleteServerById(ctx context.Context, req *pb.DeleteServerByIdRequest) (*emptypb.Empty, error) {
	s.l.Log(
		logger.INFO,
		LogMessageServer{
			"Action": "Invoked delete server",
			"Id":     req.GetId(),
		},
	)
	// Authorize

	// TO-DO : Write codo to Authorize

	//validate data
	if err := req.Validate(); err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Delete server",
				"Error":  "Invalid data in request",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// check server already exsits
	_, err := s.ServerRepo.FindOneById(int(req.GetId()))
	if err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Delete server",
				"Error":  "Can't delete because not found",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Get header
	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Update Server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
	}
	if v, ok := header["id"]; ok {
		s.ServerRepo.UpdateOneById(int(req.GetId()), map[string]interface{}{"DeletedBy": v[0]})
	}

	// Delete server
	err = s.ServerRepo.DeleteOneById(int(req.GetId()))
	if err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Delete server",
				"Error":  "Can't delete server",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return nil, nil
}

func (s *ServerService) DeleteServerByName(
	ctx context.Context,
	req *pb.DeleteServerByNameRequest,
) (*emptypb.Empty, error) {
	s.l.Log(
		logger.INFO,
		LogMessageServer{
			"Action": "Invoked delete server",
			"Name":   req.GetName(),
		},
	)
	// Authorize

	// TO-DO : Write codo to Authorize

	//validate data
	if err := req.Validate(); err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Delete server",
				"Error":  "Invalid data in request",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// check Server already exsits
	_, err := s.ServerRepo.FindOneByName(req.GetName())
	if err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Delete server",
				"Error":  "Can't delete because not found",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Get header
	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Update Server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
	}
	if v, ok := header["id"]; ok {
		s.ServerRepo.UpdateOneByName(req.GetName(), map[string]interface{}{"DeletedBy": v[0]})
	}

	// Delete Server
	err = s.ServerRepo.DeleteOneByName(req.GetName())
	if err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Delete server",
				"Error":  "Can't delete server",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return nil, nil
}
func (s *ServerService) ExportServer(ctx context.Context, req *pb.ExportServerRequest) (*emptypb.Empty, error) {
	// TO-DO code
	return nil, nil
}
func (s *ServerService) GetServerById(ctx context.Context, req *pb.GetServerByIdRequest) (*pb.Server, error) {
	s.l.Log(
		logger.INFO,
		LogMessageServer{
			"Action": "Invoked get server by id",
			"Id":     req.GetId(),
		},
	)

	// Authorize

	// TO-DO : Write codo to Authorize

	//validate data
	if err := req.Validate(); err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Get Server by id",
				"Error":  err,
			},
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get Server
	server, err := s.ServerRepo.FindOneById(int(req.GetId()))
	if err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Get Server",
				"Error":  "Internal Server",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return ConvertServerModelToServerProto(*server), nil
}
func (s *ServerService) GetServerByName(ctx context.Context, req *pb.GetServerByNameRequest) (*pb.Server, error) {
	s.l.Log(
		logger.INFO,
		LogMessageServer{
			"Action": "Invoked get server by name",
			"Name":   req.GetName(),
		},
	)

	// Authorize

	// TO-DO : Write codo to Authorize

	//validate data
	if err := req.Validate(); err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Get Server by name",
				"Error":  err,
			},
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get Server
	server, err := s.ServerRepo.FindOneByName(req.GetName())
	if err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Get Server by name",
				"Error":  "Internal Server",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return ConvertServerModelToServerProto(*server), nil
}
func (s *ServerService) ImportServer(pb.ServerService_ImportServerServer) error {
	// TO-DO code
	return nil
}
func (s *ServerService) ListServers(ctx context.Context, req *pb.ListServerRequest) (*pb.ListServersResponse, error) {
	s.l.Log(
		logger.INFO,
		LogMessageServer{
			"Action": "Invoked list server",
		},
	)
	// Authorize

	// TO-DO : Write codo to Authorize

	//validate data
	if err := req.Validate(); err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "List User",
				"Error":  "Invalid data in request",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get server
	servers, err := s.ServerRepo.FindServers(&user.FilterAdapter{Filter: req.GetFilter()})
	if err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "List server",
				"Error":  err,
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.ListServersResponse{Servers: ConvertListServerModelToListServerProto(servers)}, nil
}
func (s *ServerService) UpdateServer(ctx context.Context, req *pb.UpdateServerRequest) (*pb.Server, error) {
	s.l.Log(
		logger.INFO,
		LogMessageServer{
			"Action": "Invoked update server",
			"Id":     req.GetId(),
		},
	)

	// TO-DO Authorize

	// validate data
	if err := req.Validate(); err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Create server",
				"Id":     req.GetId(),
				"Error":  "Bad request",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// check already exists server
	if existsServer, _ := s.ServerRepo.FindOneByName(req.GetName()); existsServer != nil &&
		existsServer.ID != uint(req.GetId()) {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Update server",
				"Id":     req.GetId(),
				"Error":  "Already Exists",
				"Detail": fmt.Errorf("Already Exists server name", req.GetName()),
			},
		)
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Already Exists server name", req.GetName()))
	}

	// Parse data request
	server, err := ParseMapUpdateServerRequest(req)
	if err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Update server",
				"Id":     req.GetId(),
				"Error":  "Parse data",
				"Detail": fmt.Errorf("Parse data", err),
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Get header
	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Update Server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
	}

	if v, ok := header["id"]; ok {
		server["UpdatedBy"] = v[0]
	}

	// update server
	updatedServer, err := s.ServerRepo.UpdateOneById(int(req.GetId()), server)
	if err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Update server",
				"Id":     req.GetId(),
				"Error":  "Can't update server",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return ConvertServerModelToServerProto(*updatedServer), nil
}

func ConvertStatusServerModelToStatusServerProto(server model.ServerStatus) pb.ServerStatus {
	switch server {
	case model.On:
		return pb.ServerStatus_ON
	case model.Off:
		return pb.ServerStatus_OFF
	default:
		return pb.ServerStatus_STATUSNONE
	}
}

func ConvertServerModelToServerProto(server model.Server) *pb.Server {
	return &pb.Server{
		Id:        int64(server.ID),
		CreatedAt: server.CreatedAt.String(),
		CreatedBy: int64(server.CreatedBy),
		UpdatedAt: server.UpdatedAt.String(),
		UpdatedBy: int64(server.UpdatedBy),
		Name:      server.Name,
		Status:    ConvertStatusServerModelToStatusServerProto(server.Status),
		Ipv4:      server.Ipv4,
	}
}

func ConvertListServerModelToListServerProto(s []model.Server) []*server.Server {
	var result []*server.Server = make([]*server.Server, 0)
	for _, v := range s {
		result = append(result, ConvertServerModelToServerProto(v))
	}
	return result
}
