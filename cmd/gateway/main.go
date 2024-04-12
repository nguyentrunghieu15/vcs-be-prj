package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/auth"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	gateway_middleware "github.com/nguyentrunghieu15/vcs-be-prj/pkg/gateway/middleware"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/logger"
	authpb "github.com/nguyentrunghieu15/vcs-common-prj/apu/auth"
	serverpb "github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	userpb "github.com/nguyentrunghieu15/vcs-common-prj/apu/user"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/managedb"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
)

func CreateSupperAdmin() {

	dsnPostgres := fmt.Sprintf(
		"host=%v user=%v password=%v dbname=%v port=%v sslmode=%v",
		env.GetEnv("POSTGRES_ADDRESS"),
		env.GetEnv("POSTGRES_USERNAME"),
		env.GetEnv("POSTGRES_PASSWORD"),
		env.GetEnv("POSTGRES_DATABASE"),
		env.GetEnv("POSTGRES_PORT"),
		env.GetEnv("POSTGRES_SSLMODE"),
	)

	var (
		SupperAdmin = map[string]string{
			"Email":    env.GetEnv("ADMIN_EMAIL").(string),
			"Password": env.GetEnv("ADMIN_PASSWORD").(string),
			"FullName": env.GetEnv("ADMIN_FULLNAME").(string),
		}
	)

	postgres, err := managedb.GetConnection(
		managedb.Connection{
			Context: &managedb.PostgreContext{},
			Dsn:     dsnPostgres,
		})

	if err != nil {
		log.Fatalf("AuthService : Can't connect to PostgresSQL Database :%v", err)
	}
	log.Println("Auth Services: Connected database")
	connPostgres, _ := postgres.(*gorm.DB)
	hashPassword, err := (&auth.BcryptService{}).HashPassword(SupperAdmin["Password"])
	if err != nil {
		log.Fatalln("Can't hash password of supper admin")
	}
	_, err = model.CreateUserRepository(connPostgres).CreateUser(
		map[string]interface{}{
			"Email":         SupperAdmin["Email"],
			"Password":      hashPassword,
			"FullName":      SupperAdmin["FullName"],
			"IsSupperAdmin": true,
			"Roles":         model.RoleAdmin,
		})
	if err != nil {
		log.Fatalln("Can't create supper admin", err)
	}
}

func main() {
	gatewayConfigEnv := map[string]env.ConfigEnv{
		"JWT_SECRETKEY": {IsRequire: true, Type: env.STRING},

		"ADMIN_EMAIL":    {IsRequire: true, Type: env.STRING},
		"ADMIN_PASSWORD": {IsRequire: true, Type: env.STRING},
		"ADMIN_FULLNAME": {IsRequire: true, Type: env.STRING},

		"GATEWAY_PORT":          {IsRequire: true, Type: env.INT},
		"GATEWAY_ADDRESS":       {IsRequire: true, Type: env.STRING},
		"GATEWAY_LOG_PATH":      {IsRequire: true, Type: env.STRING},
		"GATEWAY_NAME_FILE_LOG": {IsRequire: true, Type: env.STRING},

		"AUTH_PORT":    {IsRequire: true, Type: env.INT},
		"AUTH_ADDRESS": {IsRequire: true, Type: env.STRING},

		"USER_PORT":    {IsRequire: true, Type: env.INT},
		"USER_ADDRESS": {IsRequire: true, Type: env.STRING},

		"SERVER_PORT":    {IsRequire: true, Type: env.INT},
		"SERVER_ADDRESS": {IsRequire: true, Type: env.STRING},

		"POSTGRES_ADDRESS":  {IsRequire: true, Type: env.STRING},
		"POSTGRES_PORT":     {IsRequire: true, Type: env.INT},
		"POSTGRES_USERNAME": {IsRequire: true, Type: env.STRING},
		"POSTGRES_PASSWORD": {IsRequire: true, Type: env.STRING},
		"POSTGRES_DATABASE": {IsRequire: true, Type: env.STRING},
		"POSTGRES_SSLMODE":  {IsRequire: true, Type: env.STRING},
	}
	env.Load(".env", gatewayConfigEnv)
	e := echo.New()
	newLogger := logger.NewLogger()
	newLogger.Config = logger.LoggerConfig{
		IsLogRotate:     true,
		PathToLog:       env.GetEnv("GATEWAY_LOG_PATH").(string),
		FileNameLogBase: env.GetEnv("GATEWAY_NAME_FILE_LOG").(string),
	}
	e.Use(newLogger.ImplementedMiddlewareLogger())
	e.Use(middleware.Recover())
	e.Static("/static", "static")

	createAdmin := flag.Bool("create-admin", false, "Create a supperadmin account")

	flag.Parse()
	if *createAdmin {
		CreateSupperAdmin()
	}

	mux := runtime.NewServeMux()
	//...

	err := authpb.RegisterAuthServiceHandlerFromEndpoint(
		context.Background(),
		mux,
		fmt.Sprintf("%v:%v", env.GetEnv("AUTH_ADDRESS"), env.GetEnv("AUTH_PORT")),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	)

	if err != nil {
		log.Fatalln("Can't connect to Auth service")
	}

	e.Any(
		"/api/v1/auth*",
		echo.WrapHandler(mux),
	)

	err = userpb.RegisterUserServiceHandlerFromEndpoint(
		context.Background(),
		mux,
		fmt.Sprintf("%v:%v", env.GetEnv("USER_ADDRESS"), env.GetEnv("USER_PORT")),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})

	if err != nil {
		log.Fatalln("Can't connect to User service")
	}
	e.Any(
		"/api/v1/user*",
		echo.WrapHandler(mux),
		gateway_middleware.UseJwtMiddleware(),
		gateway_middleware.UserParseJWTTokenMiddleware(),
	)

	err = serverpb.RegisterServerServiceHandlerFromEndpoint(
		context.Background(),
		mux,
		fmt.Sprintf("%v:%v", env.GetEnv("SERVER_ADDRESS"), env.GetEnv("SERVER_PORT")),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})

	if err != nil {
		log.Fatalln("Can't connect to Server service")
	}
	e.Any(
		"/api/v1/server*",
		echo.WrapHandler(mux),
		gateway_middleware.UseJwtMiddleware(),
		gateway_middleware.UserParseJWTTokenMiddleware(),
	)

	e.Logger.Fatal(e.Start(fmt.Sprintf("%v:%v", env.GetEnv("GATEWAY_ADDRESS"), env.GetEnv("GATEWAY_PORT"))))
}
