package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/auth"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/logger"
	gedis "github.com/nguyentrunghieu15/vcs-be-prj/pkg/redis"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/managedb"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"github.com/segmentio/kafka-go"
	"github.com/xuri/excelize/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type LogMessageServer map[string]interface{}

type ServerRepo interface {
	CheckServerExists(map[string]interface{}) bool
	CountServers(*string, *pb.FilterServer) (int64, error)
	CreateBacth(uint64, []map[string]interface{}) (*pb.ImportServerResponse, error)
	CreateServer(map[string]interface{}) (*model.Server, error)
	DeleteOneById(uuid.UUID) error
	DeleteOneByName(string) error
	FindOneById(uuid.UUID) (*model.Server, error)
	FindOneByName(string) (*model.Server, error)
	FindServers(*pb.ListServerRequest) ([]model.Server, error)
	UpdateOneById(uuid.UUID, map[string]interface{}) (*model.Server, error)
	UpdateOneByName(string, map[string]interface{}) (*model.Server, error)
}

type ServerService struct {
	pb.ServerServiceServer
	l          *logger.LoggerDecorator
	ServerRepo ServerRepo
	kafka      ProducerClientInterface
	auhthorize *auth.Authorizer
}

type ServerServiceKafkaLogger struct {
	l *logger.LoggerDecorator
}

func (s *ServerServiceKafkaLogger) Printf(msg string, a ...interface{}) {
	s.l.Log(logger.INFO, LogMessageServer{
		"Kafka Log": "Infor",
		"Message":   msg,
		"Details":   a,
	})
}

type ServerServiceKafkaLoggerError struct {
	l *logger.LoggerDecorator
}

func (s *ServerServiceKafkaLoggerError) Printf(msg string, a ...interface{}) {
	s.l.Log(logger.ERROR, LogMessageServer{
		"Kafka Log": "Error",
		"Message":   msg,
		"Details":   a,
	})
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
	connPostgres.Config.Logger = gormLogger.Default.LogMode(gormLogger.Info)
	newLogger := logger.NewLogger()
	newLogger.Config = logger.LoggerConfig{
		IsLogRotate:     true,
		PathToLog:       env.GetEnv("SERVER_LOG_PATH").(string),
		FileNameLogBase: env.GetEnv("SERVER_NAME_FILE_LOG").(string),
	}

	newKafka := NewProducer(ConfigProducer{
		Broker:     []string{env.GetEnv("KAFKA_BOOTSTRAP_SERVER").(string)},
		Topic:      env.GetEnv("KAFKA_TOPIC_EXPORT").(string),
		Balancer:   &kafka.LeastBytes{},
		Logger:     &ServerServiceKafkaLogger{l: newLogger},
		ErrorLoger: &ServerServiceKafkaLoggerError{l: newLogger},
	})

	newRedisConfig := gedis.GedisConfig{
		Addess:   fmt.Sprintf("%v:%v", env.GetEnv("REDIS_HOST"), env.GetEnv("REDIS_PORT")),
		Password: env.GetEnv("REDIS_PASSWORD").(string),
		Username: env.GetEnv("REDIS_USERNAME").(string),
	}

	return &ServerService{
		ServerRepo: NewServerRepoProxy(newRedisConfig, connPostgres),
		l:          newLogger,
		kafka:      newKafka,
		auhthorize: &auth.Authorizer{},
	}
}

func (s *ServerService) CreateServer(ctx context.Context, req *pb.CreateServerRequest) (*pb.Server, error) {
	s.l.Log(
		logger.INFO,
		LogMessageServer{
			"Action": "Invoked create server",
			"Name":   req.GetName(),
		},
	)

	// Authorize
	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Create server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	role, ok := header["role"]
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Create server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	if !s.auhthorize.HavePermisionToCreateServer(model.UserRole(role[0])) {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Create server",
				"Error":  "Permission denie",
			},
		)
		return nil, status.Error(codes.PermissionDenied, "Can't create user")
	}

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
		return nil, status.Error(codes.AlreadyExists,
			fmt.Sprintf("Already Exists server name", req.GetName()))
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
	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Delete server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	role, ok := header["role"]
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Delete server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	if !s.auhthorize.HavePermisionToDeleteServer(model.UserRole(role[0])) {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Delete server",
				"Error":  "Permission denie",
			},
		)
		return nil, status.Error(codes.PermissionDenied, "Can't delete user")
	}

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
	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Delete server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	role, ok := header["role"]
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Delete server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	if !s.auhthorize.HavePermisionToDeleteServer(model.UserRole(role[0])) {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Delete server",
				"Error":  "Permission denie",
			},
		)
		return nil, status.Error(codes.PermissionDenied, "Can't delete user")
	}

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
	s.l.Log(
		logger.INFO,
		LogMessageServer{
			"Action": "Invoked Export server",
		},
	)
	// Authorize

	// Authorize
	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Export server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	role, ok := header["role"]
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Export server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	if !s.auhthorize.HavePermisionToExportServer(model.UserRole(role[0])) {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Export server",
				"Error":  "Permission denie",
			},
		)
		return nil, status.Error(codes.PermissionDenied, "Can't export user")
	}

	// TO-DO : Write codo to Authorize

	//validate data
	if err := req.Validate(); err != nil {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Export server",
				"Error":  "Invalid data in request",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	//Parse to kafka message
	parseMessage, err := ParseExportRequestToKafkaMessage(req)
	if err != nil {
		s.l.Log(logger.ERROR, LogMessageServer{
			"Action": "Export",
			"Error":  err,
		})
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Write to kafka
	err = s.kafka.WriteMessage(context.Background(), *parseMessage)
	if err != nil {
		return nil, status.Error(codes.Aborted, "Can't export")
	}
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
	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Get server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	role, ok := header["role"]
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Get server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	if !s.auhthorize.HavePermisionToViewServer(model.UserRole(role[0])) {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Get server",
				"Error":  "Permission denie",
			},
		)
		return nil, status.Error(codes.PermissionDenied, "Can't get user")
	}

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
	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Get server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	role, ok := header["role"]
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Get server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	if !s.auhthorize.HavePermisionToViewServer(model.UserRole(role[0])) {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Get server",
				"Error":  "Permission denie",
			},
		)
		return nil, status.Error(codes.PermissionDenied, "Can't get user")
	}

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
		return status.Error(codes.Internal, err.Error())
	}

	defer newF.Close()

	for {
		p, err := stream.Recv()
		if err != nil {
			if err == io.EOF {

				newF.Close()

				f, err := excelize.OpenFile(filePath)
				if err != nil {
					s.l.Log(
						logger.ERROR,
						LogMessageServer{
							"Acction": "Import Server",
							"Error":   "Open file excel",
							"Detail":  err,
						},
					)
					return status.Error(codes.Internal, err.Error())
				}
				defer func() {
					// Close the spreadsheet.
					if err := f.Close(); err != nil {
						s.l.Log(
							logger.ERROR,
							LogMessageServer{
								"Acction": "Import Server",
								"Error":   "Close file excel",
								"Detail":  err,
							},
						)
					}
				}()

				rows, err := f.GetRows("Sheet1")
				if err != nil {
					s.l.Log(
						logger.ERROR,
						LogMessageServer{
							"Acction": "Import Server",
							"Error":   "Read file excel",
							"Detail":  err,
						},
					)
					return status.Error(codes.Internal, err.Error())
				}

				listServer := make([]map[string]interface{}, 0)

				fieldName := rows[0]
				for _, v := range fieldName {
					if v != "id" && v != "name" && v != "ipv4" && v != "status" {
						return status.Error(codes.InvalidArgument, "Invalid field name")
					}
				}

				for _, row := range rows[1:] {
					server := make(map[string]interface{})
					for idx, colCell := range row {
						server[fieldName[idx]] = colCell
					}
					isValidServer := ValidateServerFormMap(server)
					if isValidServer != nil {
						return status.Error(codes.InvalidArgument, isValidServer.Error())
					}
					listServer = append(listServer, server)
				}

				userId, _ := strconv.ParseUint(user[0], 0, 0)
				result, err := s.ServerRepo.CreateBacth(userId, listServer)
				if err != nil {
					return status.Error(codes.Internal, err.Error())
				}

				stream.SendAndClose(result)
				break
			}
			return status.Error(codes.InvalidArgument, err.Error())
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
	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "List server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	role, ok := header["role"]
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "List server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	if !s.auhthorize.HavePermisionToViewServer(model.UserRole(role[0])) {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "List server",
				"Error":  "Permission denie",
			},
		)
		return nil, status.Error(codes.PermissionDenied, "Can't list user")
	}

	// TO-DO : Write codo to Authorize

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

	// Authorize
	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Update server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	role, ok := header["role"]
	if !ok {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Update server",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	if !s.auhthorize.HavePermisionToUpdateServer(model.UserRole(role[0])) {
		s.l.Log(
			logger.ERROR,
			LogMessageServer{
				"Action": "Update server",
				"Error":  "Permission denie",
			},
		)
		return nil, status.Error(codes.PermissionDenied, "Can't update user")
	}

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
