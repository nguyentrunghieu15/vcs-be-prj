package user

import (
	"context"
	"fmt"
	"log"

	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	"github.com/nguyentrunghieu15/vcs-common-prj/apu/user"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/managedb"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type UserServer struct {
	user.UserServiceServer
	UserRepo *UserRepositoryDecorator
}

func NewUserServer() UserServer {
	dsnPostgres := fmt.Sprintf(
		"host=%v user=%v password=%v dbname=%v port=%v sslmode=%v",
		env.GetEnv("POSTGRES_ADDRESS"),
		env.GetEnv("POSTGRES_USERNAME"),
		env.GetEnv("POSTGRES_PASSWORD"),
		env.GetEnv("POSTGRES_DATABASE"),
		env.GetEnv("POSTGRES_PORT"),
		env.GetEnv("POSTGRES_SSLMODE"),
	)
	postgres, err := managedb.GetConnection(
		managedb.Connection{
			Context: &managedb.PostgreContext{},
			Dsn:     dsnPostgres,
		})
	if err != nil {
		log.Fatalf("AuthService : Can't connect to PostgresSQL Database :%v", err)
	}
	log.Println("Connected database")
	connPostgres, _ := postgres.(*gorm.DB)
	connPostgres.Config.Logger = gormLogger.Default.LogMode(gormLogger.Silent)

	return UserServer{UserRepo: NewUserRepository(connPostgres)}
}

func (u *UserServer) ListUsers(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	return nil, nil
}
func (u *UserServer) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.User, error) {
	return nil, nil
}
func (u *UserServer) UpdateUser(ctx context.Context, req *user.UpdateUserByIdRequest) (*user.User, error) {
	return nil, nil
}
func (u *UserServer) DeleteUSer(ctx context.Context, req *user.DeleteUserByIdRequest) (*emptypb.Empty, error) {
	return nil, nil
}
