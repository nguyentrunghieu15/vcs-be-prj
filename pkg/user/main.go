package user

import (
	"context"
	"fmt"
	"log"

	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/auth"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/logger"
	"github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/apu/user"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/managedb"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type UserServer struct {
	user.UserServiceServer
	UserRepo *UserRepositoryDecorator
	l        *logger.LoggerDecorator
	bcrypt   *auth.BcryptService
}

// GetUser implements user.UserServiceServer.
func (u UserServer) GetUser(ctx context.Context, req *user.GetUserByIdRequest) (*user.User, error) {
	u.l.Log(
		logger.ERROR,
		LogMessageUser{
			"Action": "Invoked get user",
			"Id":     req.GetId(),
		},
	)
	// Authorize

	// TO-DO : Write codo to Authorize

	// Get user
	user, err := u.UserRepo.FindOneById(int(req.GetId()))
	if err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Get User",
				"Error":  err,
			},
		)
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return ConvertUserModelToUserProto(*user), nil
}

// ListUsers implements user.UserServiceServer.
func (u *UserServer) ListUsers(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	u.l.Log(
		logger.ERROR,
		LogMessageUser{
			"Action": "Invoked list user",
		},
	)
	// Authorize

	// TO-DO : Write codo to Authorize

	// Get user
	users, err := u.UserRepo.FindUsers(&FilterAdapter{Filter: req.GetFilter()})
	if err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "List User",
				"Error":  err,
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &user.ListUsersResponse{Users: ConvertListUserModelToListUserProto(users)}, nil
}

// CreateUser implements user.UserServiceServer.
func (u *UserServer) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.User, error) {
	u.l.Log(
		logger.ERROR,
		LogMessageUser{
			"Action": "Invoked create user",
			"Email":  req.GetEmail(),
		},
	)
	// Authorize

	// TO-DO : Write codo to Authorize

	// check user already exsits
	exsitsUser, _ := u.UserRepo.FindOneByEmail(req.GetEmail())
	if exsitsUser != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Create User",
				"Error":  "Can't create because user already exisit",
			},
		)
		return nil, status.Error(
			codes.AlreadyExists,
			fmt.Sprintf("Already exisit email:%v", req.GetEmail()),
		)
	}

	// Create user
	hashedPassword, err := u.bcrypt.HashPassword(req.GetPassword())
	if err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Create User",
				"Error":  "Can't hash password",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
	}
	var user = model.User{
		Email:         req.GetEmail(),
		FullName:      req.GetFullName(),
		Phone:         req.GetPhone(),
		Password:      hashedPassword,
		Avatar:        req.GetAvatar(),
		IsSupperAdmin: req.GetIsSupperAdmin(),
		Roles:         ConvertUserRoleProtoToUserRoleModel(req.GetRoles()),
	}
	err = u.UserRepo.CreateUser(&user)
	if err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Create User",
				"Error":  "Can't create",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	return ConvertUserModelToUserProto(user), nil
}

// UpdateUser implements user.UserServiceServer.
func (u *UserServer) UpdateUser(ctx context.Context, req *user.UpdateUserByIdRequest) (*user.User, error) {
	u.l.Log(
		logger.ERROR,
		LogMessageUser{
			"Action": "Invoked update user",
			"Id":     req.GetId(),
		},
	)
	// Authorize

	// TO-DO : Write codo to Authorize

	// check user already exsits
	_, err := u.UserRepo.FindOneById(int(req.GetId()))
	if err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Update User",
				"Error":  "Can't update because not found",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Update user
	var user = model.User{
		Email:    req.GetEmail(),
		FullName: req.GetFullName(),
		Phone:    req.GetPhone(),
		Avatar:   req.GetAvatar(),
		Roles:    ConvertUserRoleProtoToUserRoleModel(req.GetRoles()),
	}
	err = u.UserRepo.UpdateOneById(int(req.GetId()), &user)
	if err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Update User",
				"Error":  "Can't update user",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return ConvertUserModelToUserProto(user), nil
}

// DeleteUSer implements user.UserServiceServer.
func (u *UserServer) DeleteUser(ctx context.Context, req *user.DeleteUserByIdRequest) (*emptypb.Empty, error) {
	u.l.Log(
		logger.ERROR,
		LogMessageUser{
			"Action": "Invoked delete user",
			"Id":     req.GetId(),
		},
	)
	// Authorize

	// TO-DO : Write codo to Authorize

	// check user already exsits
	_, err := u.UserRepo.FindOneById(int(req.GetId()))
	if err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Delete User",
				"Error":  "Can't delete because not found",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Delete user
	err = u.UserRepo.DeleteOneById(int(req.GetId()))
	if err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Update User",
				"Error":  "Can't update user",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return nil, nil
}

func NewUserServer() *UserServer {
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
	newLogger := logger.NewLogger()
	newLogger.Config = logger.LoggerConfig{
		IsLogRotate:     true,
		PathToLog:       env.GetEnv("USER_LOG_PATH").(string),
		FileNameLogBase: env.GetEnv("USER_NAME_FILE_LOG").(string),
	}
	return &UserServer{
		UserRepo: NewUserRepository(connPostgres),
		l:        newLogger,
		bcrypt:   &auth.BcryptService{},
	}
}

type FilterAdapter struct {
	Filter *server.Filter
}

// GetLimit implements model.FilterQueryInterface.
func (f *FilterAdapter) GetLimit() int64 {
	return f.Filter.GetLimit()
}

// GetPage implements model.FilterQueryInterface.
func (f *FilterAdapter) GetPage() int64 {
	return f.Filter.GetPage()
}

// GetPageSize implements model.FilterQueryInterface.
func (f *FilterAdapter) GetPageSize() int64 {
	return f.Filter.GetPageSize()
}

// GetSort implements model.FilterQueryInterface.
func (f *FilterAdapter) GetSort() model.TypeSort {
	switch f.Filter.GetSort() {
	case server.TypeSort_ASC:
		return model.ASC
	case server.TypeSort_DESC:
		return model.DESC
	default:
		return model.NONE
	}
}

// GetSortBy implements model.FilterQueryInterface.
func (f *FilterAdapter) GetSortBy() string {
	return f.Filter.GetSortBy()
}

func ConvertUserRoleModelToUserRoleProto(role model.UserRole) user.UserRole {
	switch role {
	case model.RoleAdmin:
		return user.UserRole_RoleAdmin
	case model.RoleUser:
		return user.UserRole_RoleUser
	default:
		return user.UserRole_RoleUser
	}
}

func ConvertUserRoleProtoToUserRoleModel(role user.UserRole) model.UserRole {
	switch role {
	case user.UserRole_RoleAdmin:
		return model.RoleAdmin
	default:
		return model.RoleUser
	}
}

func ConvertUserModelToUserProto(u model.User) *user.User {
	return &user.User{
		Id:            int64(u.ID),
		CreatedAt:     u.CreatedAt.String(),
		CreatedBy:     int64(u.CreatedBy),
		UpdatedAt:     u.UpdatedAt.String(),
		UpdatedBy:     int64(u.UpdatedBy),
		Email:         u.Email,
		FullName:      u.FullName,
		Phone:         u.Phone,
		Avatar:        u.Avatar,
		IsSupperAdmin: u.IsSupperAdmin,
		Roles:         ConvertUserRoleModelToUserRoleProto(u.Roles),
	}
}

func ConvertListUserModelToListUserProto(u []model.User) []*user.User {
	var result []*user.User = make([]*user.User, 0)
	for _, v := range u {
		result = append(result, ConvertUserModelToUserProto(v))
	}
	return result
}

type LogMessageUser map[string]interface{}
