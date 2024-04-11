package server

import (
	"context"
	"fmt"
	"log"

	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/logger"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/managedb"
	"google.golang.org/grpc/codes"
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
	return nil, nil
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

	// check user already exsits
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

	// Delete user
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
func (s *ServerService) DeleteServerByName(ctx context.Context, req *pb.DeleteServerByNameRequest) (*emptypb.Empty, error) {
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

	// check user already exsits
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

	// Delete user
	err = s.ServerRepo.DeleteOneByEmail(req.GetName())
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

	// Get user
	server, err := s.ServerRepo.FindOneById(int(req.GetId()))
	if err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Get User",
				"Error":  "Invalid data in request",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return server, nil
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

	// Get user
	server, err := s.ServerRepo.FindOneByName(req.GetName())
	if err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Get Server by name",
				"Error":  "Invalid data in request",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return server, nil
}
func (s *ServerService) ImportServer(pb.ServerService_ImportServerServer) error {

	return nil
}
func (s *ServerService) ListServers(context.Context, *pb.ListServerRequest) (*pb.ListServersResponse, error) {
	return nil, nil
}
func (s *ServerService) UpdateServer(context.Context, *pb.UpdateServerRequest) (*pb.Server, error) {
	return nil, nil
}
