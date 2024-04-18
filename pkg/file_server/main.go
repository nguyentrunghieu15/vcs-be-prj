package fileserver

import (
	"io"
	"os"
	"strings"

	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	file "github.com/nguyentrunghieu15/vcs-common-prj/apu/server_file"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type FileServer struct {
	file.FileServiceServer
}

func (f *FileServer) UploadFile(stream file.FileService_UploadFileServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Errorf(codes.DataLoss, "ClientStreamingEcho: failed to get metadata")
	}

	paths, ok := md["path"]
	if !ok {
		return status.Errorf(codes.DataLoss, "ClientStreamingEcho: failed to get metadata")
	}
	path := paths[0]
	rootPath := env.GetEnv("FILE_SERVER_FOLDER").(string) + "/" + path

	splitedPath := strings.Split(rootPath, "/")
	lenPath := len(splitedPath)

	if lenPath > 1 {
		folder := strings.Join(splitedPath[:lenPath-1], "/")
		os.MkdirAll(folder, os.ModePerm)
	}

	newFile, err := os.OpenFile(rootPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return status.Error(codes.Canceled, err.Error())
	}

	defer newFile.Close()

	for {
		data, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				// TO-DO code logic when uplaod suucess
				break
			}
			return status.Error(codes.Canceled, err.Error())
		}
		newFile.Write(data.Chunk)
	}
	return nil
}
func (f *FileServer) Download(req *file.DownloadRequestMessage, stream file.FileService_DownloadServer) error {
	path := req.GetPath()
	rootPath := env.GetEnv("FILE_SERVER_FOLDER").(string) + "/" + path

	newFile, err := os.OpenFile(rootPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return status.Error(codes.Canceled, err.Error())
	}
	defer newFile.Close()

	chunk := make([]byte, 100)
	for {
		_, err := newFile.Read(chunk)
		if err != nil {
			if err == io.EOF {
				// TO-DO code logic when uplaod suucess
				break
			}
			return status.Error(codes.Canceled, err.Error())
		}
		stream.Send(&file.DownloadResponseMessage{
			Chunk: chunk,
		})
	}
	return nil
}
