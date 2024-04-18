package main

import (
	"fmt"
	"log"
	"net"

	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	fileserver "github.com/nguyentrunghieu15/vcs-be-prj/pkg/file_server"
	file "github.com/nguyentrunghieu15/vcs-common-prj/apu/server_file"
	"google.golang.org/grpc"
)

func main() {
	fileStogareEnv := map[string]env.ConfigEnv{
		"FILE_SERVER_ADDRESS": {IsRequire: true, Type: env.STRING},
		"FILE_SERVER_PORT":    {IsRequire: true, Type: env.STRING},
		"FILE_SERVER_FOLDER":  {IsRequire: true, Type: env.STRING},
	}

	env.Load(".env", fileStogareEnv)

	lis, err := net.Listen(
		"tcp",
		fmt.Sprintf("%v:%v", env.GetEnv("FILE_SERVER_ADDRESS"), env.GetEnv("FILE_SERVER_PORT")),
	)
	if err != nil {
		log.Fatalln(err)
	}
	fileServer := fileserver.FileServer{}
	s := grpc.NewServer()
	file.RegisterFileServiceServer(s, &fileServer)
	log.Println("Create GRPC file Server")
	if err = s.Serve(lis); err != nil {
		log.Fatalln(err)
	}
}
