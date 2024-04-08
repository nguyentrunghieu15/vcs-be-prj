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
	authpb "github.com/nguyentrunghieu15/vcs-common-prj/apu/auth"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/managedb"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
)

var (
	SupperAdmin = map[string]string{
		"Email":    "",
		"Password": "",
		"FullName": "",
	}
)

func CreateSupperAdmin() {

	dsnPostgres := "host=localhost user=hiro password=1 dbname=vcs_msm_db port=5432 sslmode=disable"
	postgres, err := managedb.GetConnection(managedb.Connection{Context: &managedb.PostgreContext{}, Dsn: dsnPostgres})
	if err != nil {
		log.Fatalf("AuthService : Can't connect to PostgresSQL Database :%v", err)
	}
	log.Println("Connected database", postgres)
	connPostgres, _ := postgres.(*gorm.DB)
	hashPassword, err := (&auth.BcryptService{}).HashPassword(SupperAdmin["Password"])
	fmt.Println(hashPassword)
	if err != nil {
		log.Fatalln("Can't hash password of supper admin")
	}
	err = model.CreateUserRepository(connPostgres).CreateUser(
		&model.User{
			Email:         SupperAdmin["Email"],
			Password:      hashPassword,
			FullName:      SupperAdmin["FullName"],
			IsSupperAdmin: true,
		})
	if err != nil {
		log.Fatalln("Can't create supper admin", err)
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/static", "/static")

	createAdmin := flag.Bool("create-admin", false, "Create a supperadmin account")

	flag.Parse()
	if *createAdmin {
		CreateSupperAdmin()
	}

	mux := runtime.NewServeMux()
	//...

	authpb.RegisterAuthServiceHandlerFromEndpoint(context.Background(), mux, "localhost:5000",
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})

	e.Any("/api/v1/auth/*", echo.WrapHandler(mux)) // all HTTP requests starting with `/prefix` are handled by `grpc-gateway`

	e.Logger.Fatal(e.Start(":8080"))
}
