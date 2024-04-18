package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/server"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	file "github.com/nguyentrunghieu15/vcs-common-prj/apu/server_file"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/managedb"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"github.com/xuri/excelize/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
)

type ExporterWorker struct {
	kafkaClientConfig server.ConfigConsumer
	kafkaClient       server.ConsumerClientInterface
	db                *ExporterServerRepo
	fileRepo          *model.FileRepository
	fileStorage       file.FileServiceClient
}

func NewExporterWorker() *ExporterWorker {
	kafkaClientConfig := server.ConfigConsumer{
		Broker:     []string{env.GetEnv("KAFKA_BOOTSTRAP_SERVER").(string)},
		Topic:      env.GetEnv("KAFKA_TOPIC_EXPORT").(string),
		GroupID:    env.GetEnv("KAFKA_GROUP_ID").(string),
		Logger:     log.New(log.Writer(), "Worker", log.Flags()),
		ErrorLoger: log.New(log.Writer(), "Worker", log.Flags()),
	}

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

	fileStorageAddress := fmt.Sprintf("%v:%v", env.GetEnv("FILE_SERVER_ADDRESS"),
		env.GetEnv("FILE_SERVER_PORT"))
	conn, err := grpc.DialContext(
		context.Background(),
		fileStorageAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln(err)
	}

	return &ExporterWorker{
		kafkaClientConfig: kafkaClientConfig,
		db:                NewExporterServerRepo(connPostgres),
		fileRepo:          model.NewFileRepository(connPostgres),
		fileStorage:       file.NewFileServiceClient(conn),
	}
}

func (e *ExporterWorker) Work() {
	e.kafkaClient = server.NewComsumer(e.kafkaClientConfig)
	for {
		m, err := e.kafkaClient.ReadMessage(context.Background())
		if err != nil {
			fmt.Println(e.kafkaClientConfig)
			fmt.Println("Error", err)
			break
		}

		request, err := e.parseKafkaMessageToRequestExport(m.Value)
		if err != nil {
			fmt.Println("Error", err)
		}
		go e.ExportToFileLocal(request)
	}

	if err := e.kafkaClient.CloseReader(); err != nil {
		log.Fatal("failed to stop worker:", err)
	}
}

func (e *ExporterWorker) parseKafkaMessageToRequestExport(msg []byte) (*pb.ExportServerRequest, error) {
	var req pb.ExportServerRequest
	err := json.Unmarshal(msg, &req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (e *ExporterWorker) ExportToFileLocal(req *pb.ExportServerRequest) error {
	servers, err := e.db.FindAllServer(req)
	if err != nil {
		fmt.Println(err)
	}

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	var colName = []string{"id", "name", "ipv4", "status"}
	var colCell = []string{"A", "B", "C", "D"}

	f.SetCellValue("Sheet1", "A1", colName[0])
	f.SetCellValue("Sheet1", "B1", colName[1])
	f.SetCellValue("Sheet1", "C1", colName[2])
	f.SetCellValue("Sheet1", "D1", colName[3])

	for idxServer, server := range servers {
		valueCell := []interface{}{
			server.ID,
			server.Name,
			server.Ipv4,
			server.Status,
		}
		for idxCol, cell := range colCell {
			f.SetCellValue(
				"Sheet1",
				fmt.Sprintf("%v%v", cell, idxServer+2),
				valueCell[idxCol],
			)
		}
	}

	idFile, _ := uuid.NewUUID()
	pathTemp := fmt.Sprintf("%v/%v_%v_%v",
		env.GetEnv("EXPORTER_TEMP_FOLDER"),
		req.UserId,
		idFile,
		req.File.FileName)

	if err := f.SaveAs(pathTemp); err != nil {
		fmt.Println(err)
	}

	pathServerStorage := fmt.Sprintf("%v/%v_%v", req.UserId, idFile, req.File.FileName)

	md := metadata.Pairs("path", pathServerStorage)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	e.fileRepo.CreateFile(map[string]interface{}{
		"id":        idFile,
		"file_name": req.File.FileName,
		"file_path": pathServerStorage,
		"status":    model.Exporting,
		"owner":     req.UserId,
	})

	err = e.SendToFileServer(ctx, pathTemp)
	if err != nil {
		log.Println(err)
		e.fileRepo.UpdateFileById(idFile, map[string]interface{}{"status": model.Fail})
		return err
	}
	e.fileRepo.UpdateFileById(idFile, map[string]interface{}{"status": model.Exported})
	defer os.Remove(pathTemp)
	return nil
}

func (e *ExporterWorker) SendToFileServer(ctx context.Context, path string) error {
	stream, err := e.fileStorage.UploadFile(ctx)
	if err != nil {
		log.Println("Can't send file to storage")
		return err
	}

	chunk := make([]byte, 100)

	newFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Can't read file")
		return err
	}
	defer newFile.Close()
	for {
		_, err := newFile.Read(chunk)
		if err != nil {
			if err == io.EOF {
				_, resError := stream.CloseAndRecv()
				if resError != nil && resError != io.EOF {
					log.Println("Error when complete send file")
					return resError
				}
				break
			}
			log.Println("Error when send file")
			return err
		}
		err = stream.Send(&file.UploadMessage{Chunk: chunk})
		if err != nil && err != io.EOF {
			log.Println("Error when send file")
			return err
		}
	}
	return nil
}
