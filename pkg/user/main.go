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
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type UserServer struct {
	user.UserServiceServer
	UserRepo  *UserRepositoryDecorator
	l         *logger.LoggerDecorator
	bcrypt    *auth.BcryptService
	authorize *auth.Authorizer
}

// GetUser implements user.UserServiceServer.
func (u UserServer) GetUser(ctx context.Context, req *user.GetUserByIdRequest) (*user.User, error) {
	u.l.Log(
		logger.INFO,
		LogMessageUser{
			"Action": "Invoked get user",
			"Id":     req.GetId(),
		},
	)

	// Authorize

	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Get user",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	role, ok := header["role"]
	if !ok {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Get user",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	if !u.authorize.HavePermisionToViewUser(model.UserRole(role[0])) {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Get user",
				"Error":  "Permission denie",
			},
		)
		return nil, status.Error(codes.PermissionDenied, "Can't get user")
	}

	// TO-DO : Write codo to Authorize

	//validate data
	if err := req.Validate(); err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Get User",
				"Error":  err,
			},
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get user
	user, err := u.UserRepo.FindOneById(int(req.GetId()))
	if err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Get User",
				"Error":  "Invalid data in request",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return ConvertUserModelToUserProto(*user), nil
}

// GetUserByEmail implements user.UserServiceServer.
func (u UserServer) GetUserByEmail(ctx context.Context, req *user.GetUserByEmailRequest) (*user.User, error) {
	u.l.Log(
		logger.INFO,
		LogMessageUser{
			"Action": "Invoked get user",
			"Id":     req.GetEmail(),
		},
	)

	// Authorize
	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Get user",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	role, ok := header["role"]
	if !ok {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Get user",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	if !u.authorize.HavePermisionToViewUser(model.UserRole(role[0])) {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Get user",
				"Error":  "Permission denie",
			},
		)
		return nil, status.Error(codes.PermissionDenied, "Can't get user")
	}

	// TO-DO : Write codo to Authorize

	//validate data
	if err := req.Validate(); err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Get User",
				"Error":  err,
			},
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get user
	user, err := u.UserRepo.FindOneByEmail(req.GetEmail())
	if err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Get User",
				"Error":  "Invalid data in request",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return ConvertUserModelToUserProto(*user), nil
}

// ListUsers implements user.UserServiceServer.
func (u *UserServer) ListUsers(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	u.l.Log(
		logger.INFO,
		LogMessageUser{
			"Action": "Invoked list user",
		},
	)
	// Authorize

	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Create get user",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	role, ok := header["role"]
	if !ok {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "List user",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	if !u.authorize.HavePermisionToViewUser(model.UserRole(role[0])) {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "List user",
				"Error":  "Permission denie",
			},
		)
		return nil, status.Error(codes.PermissionDenied, "Can't get user")
	}

	// TO-DO : Write codo to Authorize

	//validate data
	if err := req.Validate(); err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "List User",
				"Error":  "Invalid data in request",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := ValidateListUserQuery(req); err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "List User",
				"Error":  "Invalid data in request",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get user
	users, err := u.UserRepo.FindUsers(req)
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
		logger.INFO,
		LogMessageUser{
			"Action": "Invoked create user",
			"Email":  req.GetEmail(),
		},
	)
	// Authorize

	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Create  user",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	role, ok := header["role"]
	if !ok {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Create user",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	if !u.authorize.HavePermisionToCreateUser(model.UserRole(role[0])) {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Create user",
				"Error":  "Permission denie",
			},
		)
		return nil, status.Error(codes.PermissionDenied, "Can't create user")
	}

	// TO-DO : Write codo to Authorize

	//validate data
	if err := req.Validate(); err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Create User",
				"Email":  req.GetEmail(),
				"Error":  "Invalid data in request",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

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

	//parse data form request
	user, err := ParseMapCreateUserRequest(req)
	if err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Create User",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
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
	user["Password"] = hashedPassword

	if v, ok := header["id"]; ok {
		user["CreatedBy"] = v[0]
	}

	createdUser, err := u.UserRepo.CreateUser(user)
	if err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Create User",
				"Error":  "Can't create",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return ConvertUserModelToUserProto(*createdUser), nil
}

// UpdateUser implements user.UserServiceServer.
func (u *UserServer) UpdateUser(ctx context.Context, req *user.UpdateUserByIdRequest) (*user.User, error) {
	u.l.Log(
		logger.INFO,
		LogMessageUser{
			"Action": "Invoked update user",
			"Id":     req.GetId(),
		},
	)
	// Authorize

	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Update user",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	role, ok := header["role"]
	if !ok {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Update user",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	if !u.authorize.HavePermisionToUpdateUser(model.UserRole(role[0])) {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Update user",
				"Error":  "Permission denie",
			},
		)
		return nil, status.Error(codes.PermissionDenied, "Can't update user")
	}

	// TO-DO : Write codo to Authorize

	//validate data
	if err := req.Validate(); err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Update user",
				"Error":  "Invalid data in request",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

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

	//parse data form request
	user, err := ParseMapUpdateUserRequest(req)
	if err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Create User",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
	}

	if v, ok := header["id"]; ok {
		user["UpdatedBy"] = v[0]
	}

	// Update user
	updatedUser, err := u.UserRepo.UpdateOneById(int(req.GetId()), user)
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
	return ConvertUserModelToUserProto(*updatedUser), nil
}

// DeleteUSer implements user.UserServiceServer.
func (u *UserServer) DeleteUser(ctx context.Context, req *user.DeleteUserByIdRequest) (*emptypb.Empty, error) {
	u.l.Log(
		logger.INFO,
		LogMessageUser{
			"Action": "Invoked delete user",
			"Id":     req.GetId(),
		},
	)
	// Authorize

	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Create get user",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	role, ok := header["role"]
	if !ok {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Delete user",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	if !u.authorize.HavePermisionToDeleteUser(model.UserRole(role[0])) {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Delete user",
				"Error":  "Permission denie",
			},
		)
		return nil, status.Error(codes.PermissionDenied, "Can't delete user")
	}

	// TO-DO : Write codo to Authorize

	//validate data
	if err := req.Validate(); err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Delete user",
				"Error":  "Invalid data in request",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

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

	if v, ok := header["id"]; ok {
		u.UserRepo.UpdateOneById(int(req.GetId()), map[string]interface{}{"DeletedBy": v[0]})
	}

	// Delete user
	err = u.UserRepo.DeleteOneById(int(req.GetId()))
	if err != nil {
		u.l.Log(
			logger.ERROR,
			LogMessageUser{
				"Action": "Delete User",
				"Error":  "Can't delete user",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return nil, nil
}

func (u *UserServer) ChangePasswordUser(ctx context.Context, req *user.ChangePasswordRequest) (*user.User, error) {
	// TO-DO code
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
	connPostgres.Config.Logger = gormLogger.Default.LogMode(gormLogger.Info)
	newLogger := logger.NewLogger()
	newLogger.Config = logger.LoggerConfig{
		IsLogRotate:     true,
		PathToLog:       env.GetEnv("USER_LOG_PATH").(string),
		FileNameLogBase: env.GetEnv("USER_NAME_FILE_LOG").(string),
	}
	return &UserServer{
		UserRepo:  NewUserRepository(connPostgres),
		l:         newLogger,
		bcrypt:    &auth.BcryptService{},
		authorize: &auth.Authorizer{},
	}
}

func ConvertUserRoleModelToUserRoleProto(role model.UserRole) user.UserRole {
	switch role {
	case model.RoleAdmin:
		return user.UserRole_RoleAdmin
	case model.RoleUser:
		return user.UserRole_RoleUser
	default:
		return user.UserRole_RoleNone
	}
}

func ConvertUserRoleProtoToUserRoleModel(role user.UserRole) model.UserRole {
	switch role {
	case user.UserRole_RoleAdmin:
		return model.RoleAdmin
	case user.UserRole_RoleUser:
		return model.RoleUser
	default:
		return ""
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

func ValidateListUserQuery(req *user.ListUsersRequest) error {
	if req.GetPagination() != nil {
		if limit := req.GetPagination().Limit; limit != nil && *limit < 1 {
			return fmt.Errorf("Limit must be a positive number")
		}

		if page := req.GetPagination().Page; page != nil && *page < 1 {
			return fmt.Errorf("Page must be a positive number")
		}

		if pageSize := req.GetPagination().PageSize; pageSize != nil && *pageSize < 1 {
			return fmt.Errorf("Page size must be a positive number")
		}

		if sort := req.GetPagination().Sort; sort != nil &&
			*sort != server.TypeSort_ASC &&
			*sort != server.TypeSort_DESC &&
			*sort != server.TypeSort_NONE {
			return fmt.Errorf("Invalid type order")
		}
	}
	return nil
}
