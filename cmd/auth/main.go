package main

import (
	"fmt"
	"log"
	"net"

	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/auth"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/auth"
	"google.golang.org/grpc"
)

func main() {
	authEnv := map[string]env.ConfigEnv{
		"JWT_SECRETKEY":      {IsRequire: true, Type: env.STRING},
		"JWT_EXPIRE_TIME":    {IsRequire: true, Type: env.INT},
		"AUTH_PORT":          {IsRequire: true, Type: env.INT},
		"AUTH_ADDRESS":       {IsRequire: true, Type: env.STRING},
		"AUTH_LOG_PATH":      {IsRequire: true, Type: env.STRING},
		"AUTH_NAME_FILE_LOG": {IsRequire: true, Type: env.STRING},

		"POSTGRES_ADDRESS":  {IsRequire: true, Type: env.STRING},
		"POSTGRES_PORT":     {IsRequire: true, Type: env.INT},
		"POSTGRES_USERNAME": {IsRequire: true, Type: env.STRING},
		"POSTGRES_PASSWORD": {IsRequire: true, Type: env.STRING},
		"POSTGRES_DATABASE": {IsRequire: true, Type: env.STRING},
		"POSTGRES_SSLMODE":  {IsRequire: true, Type: env.STRING},
	}

	env.Load("./cmd/auth/.env", authEnv)

	var Address = fmt.Sprintf("%v:%v",
		env.GetEnv("AUTH_ADDRESS"),
		env.GetEnv("AUTH_PORT"))

	// Create a listener on TCP port
	lis, err := net.Listen("tcp", Address)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer()
	authServer := auth.NewAuthServer()

	// Attach the Greeter service to the server
	pb.RegisterAuthServiceServer(s, authServer)
	// Serve gRPC Server
	log.Println("Serving gRPC on ", Address)
	log.Fatal(s.Serve(lis))

}
