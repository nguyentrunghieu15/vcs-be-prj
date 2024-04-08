package main

import (
	"log"
	"net"

	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/auth"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/auth"
	"google.golang.org/grpc"
)

const (
	Network = "tcp"
	Address = "localhost:5000"
)

func main() {

	// Create a listener on TCP port
	lis, err := net.Listen(Network, Address)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer()
	authServer := auth.CreateAuthServer()

	// Attach the Greeter service to the server
	pb.RegisterAuthServiceServer(s, authServer)
	// Serve gRPC Server
	log.Println("Serving gRPC on ", Address)
	log.Fatal(s.Serve(lis))

}
