package main

import (
	"fmt"
	"log"
	"net"

	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/server"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"google.golang.org/grpc"
)

func main() {
	serverEnv := map[string]env.ConfigEnv{
		"AUTH_PORT":    {IsRequire: true, Type: env.INT},
		"AUTH_ADDRESS": {IsRequire: true, Type: env.STRING},

		"SERVER_PORT":          {IsRequire: true, Type: env.INT},
		"SERVER_ADDRESS":       {IsRequire: true, Type: env.STRING},
		"SERVER_LOG_PATH":      {IsRequire: true, Type: env.STRING},
		"SERVER_NAME_FILE_LOG": {IsRequire: true, Type: env.STRING},

		"POSTGRES_ADDRESS":  {IsRequire: true, Type: env.STRING},
		"POSTGRES_PORT":     {IsRequire: true, Type: env.INT},
		"POSTGRES_USERNAME": {IsRequire: true, Type: env.STRING},
		"POSTGRES_PASSWORD": {IsRequire: true, Type: env.STRING},
		"POSTGRES_DATABASE": {IsRequire: true, Type: env.STRING},
		"POSTGRES_SSLMODE":  {IsRequire: true, Type: env.STRING},

		"SERVER_UPLOAD_FOLDER": {IsRequire: true, Type: env.STRING},

		"KAFKA_BOOTSTRAP_SERVER": {IsRequire: true, Type: env.STRING},
		"KAFKA_GROUP_ID":         {IsRequire: true, Type: env.STRING},
		"KAFKA_TOPIC_EXPORT":     {IsRequire: true, Type: env.STRING},
	}

	env.Load(".env", serverEnv)

	var Address = fmt.Sprintf("%v:%v",
		env.GetEnv("SERVER_ADDRESS"),
		env.GetEnv("SERVER_PORT"))

	// Create a listener on TCP port
	lis, err := net.Listen("tcp", Address)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer()
	serverService := server.NewServerService()

	// Attach the Greeter service to the server
	pb.RegisterServerServiceServer(s, serverService)
	// Serve gRPC Server
	log.Println("Serving gRPC on ", Address)
	log.Fatal(s.Serve(lis))

}
