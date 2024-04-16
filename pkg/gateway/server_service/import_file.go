package serverservice

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
)

type ServerService struct {
	client server.ServerServiceClient
}

func NewServerService(ctx context.Context, endpoint string, opts []grpc.DialOption) (*ServerService, error) {
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
	return &ServerService{client: server.NewServerServiceClient(conn)}, nil
}

func (s *ServerService) Import(c echo.Context) error {
	// Source
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	userId := c.Request().Header.Get("Grpc-Metadata-id")
	md := metadata.Pairs("filename", file.Filename, "user", userId)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	stream, err := s.client.ImportServer(ctx)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			nil,
		)
	}

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, src); err != nil {
		return err
	}

	stream.Send(&server.ImportServerRequest{Chunk: buf.Bytes()})
	result, err := stream.CloseAndRecv()
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			nil,
		)
	}

	return c.JSON(
		http.StatusOK,
		result,
	)
}
