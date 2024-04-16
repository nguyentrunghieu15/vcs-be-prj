package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/logger"
	"github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/managedb"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"github.com/xuri/excelize/v2"
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
		log.Fatalf("Server service : Can't connect to PostgresSQL Database :%v", err)
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
	id, _ := uuid.Parse(req.GetId())
	_, err := s.ServerRepo.FindOneById(id)
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
		s.ServerRepo.UpdateOneById(id, map[string]interface{}{"DeletedBy": v[0]})
	}

	// Delete server
	err = s.ServerRepo.DeleteOneById(id)
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
	id, _ := uuid.Parse(req.GetId())
	server, err := s.ServerRepo.FindOneById(id)
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
func (s *ServerService) ImportServer(stream pb.ServerService_ImportServerServer) error {
	// TO-DO code
	// Read metadata from client.
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Errorf(codes.DataLoss, "ClientStreamingEcho: failed to get metadata")
	}

	fileName, _ := md["filename"]
	user, _ := md["user"]

	filePath := fmt.Sprintf("%v/%v_%v", env.GetEnv("SERVER_UPLOAD_FOLDER"), user[0], fileName[0])
	newF, err := os.OpenFile(
		filePath,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		0666,
	)

	if err != nil {
		return err
	}

	defer newF.Close()

	for {
		p, err := stream.Recv()
		if err != nil {
			if err == io.EOF {

				newF.Close()

				f, err := excelize.OpenFile(filePath)
				if err != nil {
					fmt.Println(err)
					return err
				}
				defer func() {
					// Close the spreadsheet.
					if err := f.Close(); err != nil {
						fmt.Println(err)
					}
				}()

				rows, err := f.GetRows("Sheet1")
				if err != nil {
					fmt.Println(err)
					return err
				}

				listServer := make([]map[string]interface{}, 0)

				fieldName := rows[0]
				for _, v := range fieldName {
					if v != "id" && v != "name" && v != "ipv4" && v != "status" {
						return fmt.Errorf("Invalid field name")
					}
				}

				for _, row := range rows[1:] {
					server := make(map[string]interface{})
					for idx, colCell := range row {
						server[fieldName[idx]] = colCell
					}
					listServer = append(listServer, server)
				}

				fmt.Println(listServer)

				stream.SendAndClose(&pb.ImportServerResponse{})
				break
			}
			return err
		} else {
			newF.Write(p.Chunk)
		}
	}

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

	fmt.Println(req)
	//validate data
	if err := req.Validate(); err != nil {

		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "List server",
				"Error":  "Invalid data in request",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := ValidateListServerQuery(req); err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "List server",
				"Error":  "Invalid data in request",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get server
	servers, err := s.ServerRepo.FindServers(req)
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
	total, err := s.ServerRepo.CountServers(req.Query, req.GetFilter())
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

	return &pb.ListServersResponse{Servers: ConvertListServerModelToListServerProto(servers), Total: total}, nil
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

	id, _ := uuid.Parse(req.GetId())

	// check already exists server
	if existsServer, _ := s.ServerRepo.FindOneByName(req.GetName()); existsServer != nil &&
		existsServer.ID != id {
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
	updatedServer, err := s.ServerRepo.UpdateOneById(id, server)
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
		Id:        server.ID.String(),
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

func ValidateListServerQuery(req *server.ListServerRequest) error {
	if req.GetPagination() != nil {
		if limit := req.GetPagination().Limit; limit != nil && *limit < 1 {
			return fmt.Errorf("Limit must be a positive number")
		}

		if page := req.GetPagination().Page; page != nil && *page < 1 {
			return fmt.Errorf("Page must be a positive number")
		}

		if pageSize := req.GetPagination().PageSize; pageSize != nil && *pageSize < 1 {
			return fmt.Errorf("Page size must be a positive number")
		}

		if sort := req.GetPagination().Sort; sort != nil &&
			*sort != server.TypeSort_ASC &&
			*sort != server.TypeSort_DESC &&
			*sort != server.TypeSort_NONE {
			return fmt.Errorf("Invalid type order")
		}
	}
	return nil
}
