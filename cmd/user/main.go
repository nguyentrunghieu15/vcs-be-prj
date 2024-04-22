package main

import (
	"fmt"
	"log"
	"net"

	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/user"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/user"
	"google.golang.org/grpc"
)

func main() {
	userEnv := map[string]env.ConfigEnv{
		"AUTH_PORT":    {IsRequire: true, Type: env.INT},
		"AUTH_ADDRESS": {IsRequire: true, Type: env.STRING},

		"USER_PORT":          {IsRequire: true, Type: env.INT},
		"USER_ADDRESS":       {IsRequire: true, Type: env.STRING},
		"USER_LOG_PATH":      {IsRequire: true, Type: env.STRING},
		"USER_NAME_FILE_LOG": {IsRequire: true, Type: env.STRING},

		"POSTGRES_ADDRESS":  {IsRequire: true, Type: env.STRING},
		"POSTGRES_PORT":     {IsRequire: true, Type: env.INT},
		"POSTGRES_USERNAME": {IsRequire: true, Type: env.STRING},
		"POSTGRES_PASSWORD": {IsRequire: true, Type: env.STRING},
		"POSTGRES_DATABASE": {IsRequire: true, Type: env.STRING},
		"POSTGRES_SSLMODE":  {IsRequire: true, Type: env.STRING},
	}

	env.Load("./cmd/user/.env", userEnv)

	var Address = fmt.Sprintf("%v:%v",
		env.GetEnv("USER_ADDRESS"),
		env.GetEnv("USER_PORT"))

	// Create a listener on TCP port
	lis, err := net.Listen("tcp", Address)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer()
	userServer := user.NewUserServer()

	// Attach the Greeter service to the server
	pb.RegisterUserServiceServer(s, userServer)
	// Serve gRPC Server
	log.Println("Serving gRPC on ", Address)
	log.Fatal(s.Serve(lis))

}
