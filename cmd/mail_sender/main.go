package main

import (
	"fmt"
	"log"
	"net"

	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	mailsenderservice "github.com/nguyentrunghieu15/vcs-be-prj/pkg/mail_sender"
	"github.com/nguyentrunghieu15/vcs-common-prj/apu/mail_sender"
	"google.golang.org/grpc"
)

func main() {
	mailEnv := map[string]env.ConfigEnv{
		"POSTGRES_ADDRESS":  {IsRequire: true, Type: env.STRING},
		"POSTGRES_PORT":     {IsRequire: true, Type: env.INT},
		"POSTGRES_USERNAME": {IsRequire: true, Type: env.STRING},
		"POSTGRES_PASSWORD": {IsRequire: true, Type: env.STRING},
		"POSTGRES_DATABASE": {IsRequire: true, Type: env.STRING},
		"POSTGRES_SSLMODE":  {IsRequire: true, Type: env.STRING},

		"ELASTIC_CERT_PATH":      {IsRequire: true, Type: env.STRING},
		"ELASTICSEARCH_USERNAME": {IsRequire: true, Type: env.STRING},
		"ELASTICSEARCH_PASSWORD": {IsRequire: true, Type: env.STRING},
		"ELASTIC_ADDRESS":        {IsRequire: true, Type: env.STRING},
		"ELASTIC_CERT_FINGER":    {IsRequire: true, Type: env.STRING},

		"MAIL_SENDER_TEMPLATE":           {IsRequire: true, Type: env.STRING},
		"MAIL_SENDER_EMAIL":              {IsRequire: true, Type: env.STRING},
		"MAIL_SENDER_PASSWORD":           {IsRequire: true, Type: env.STRING},
		"MAIL_SENDER_LOG_PATH":           {IsRequire: true, Type: env.STRING},
		"MAIL_SENDER_NAME_FILE_LOG":      {IsRequire: true, Type: env.STRING},
		"MAIL_SENDER_ADDRESS":            {IsRequire: true, Type: env.STRING},
		"MAIL_SENDER_PORT":               {IsRequire: true, Type: env.INT},
		"MAIL_SENDER_EMAIL_SUPPER_ADMIN": {IsRequire: true, Type: env.STRING},
		"MAIL_SENDER_SMTP_HOST":          {IsRequire: true, Type: env.STRING},
		"MAIL_SENDER_SMTP_PORT":          {IsRequire: true, Type: env.STRING},
	}
	env.Load("./cmd/mail_sender/.env", mailEnv)

	var Address = fmt.Sprintf("%v:%v",
		env.GetEnv("MAIL_SENDER_ADDRESS"),
		env.GetEnv("MAIL_SENDER_PORT"))

	// Create a listener on TCP port
	lis, err := net.Listen("tcp", Address)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer()
	mailSenderService := mailsenderservice.NewMailSenderServer()

	go mailSenderService.WorkDaily()

	// Attach the Greeter service to the server
	mail_sender.RegisterMailServerServer(s, mailSenderService)
	// Serve gRPC Server
	log.Println("Serving gRPC on ", Address)
	log.Fatal(s.Serve(lis))

}
