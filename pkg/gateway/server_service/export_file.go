package serverservice

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	file "github.com/nguyentrunghieu15/vcs-common-prj/apu/server_file"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/managedb"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"gorm.io/gorm"
)

type ServerStogareService struct {
	client   file.FileServiceClient
	fileRepo *model.FileRepository
}

func NewServerStogareService(
	ctx context.Context,
	endpoint string,
	opts []grpc.DialOption,
) (*ServerStogareService, error) {
	fmt.Println(endpoint)
	conn, err := grpc.DialContext(ctx, endpoint, opts...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

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
	return &ServerStogareService{
		client:   file.NewFileServiceClient(conn),
		fileRepo: model.NewFileRepository(connPostgres),
	}, nil
}

func (s *ServerStogareService) Export(c echo.Context) error {
	path := c.QueryParam("path")
	stream, err := s.client.Download(context.Background(), &file.DownloadRequestMessage{Path: path})
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	pathTemp := fmt.Sprintf("%v/%v.xlsx", env.GetEnv("GATEWAY_UPLOAD_FOLDER"), uuid.New().String())
	newFile, err := os.OpenFile(pathTemp, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	defer newFile.Close()
	defer os.Remove(pathTemp)

	for {
		data, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				// TO-DO code logic when uplaod suucess
				break
			}
			fmt.Println(err)
			return c.JSON(http.StatusInternalServerError, err)
		}
		newFile.Write(data.Chunk)
	}
	newFile.Close()
	return c.File(pathTemp)
}

func (s *ServerStogareService) GetAllFileOfUser(c echo.Context) error {
	userIdString := c.Param("id")
	userId, err := strconv.ParseInt(userIdString, 2, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": err.Error(),
		})
	}
	listFile, err := s.fileRepo.FindAllFileOfUser(int(userId))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"files": *listFile,
	})
}
