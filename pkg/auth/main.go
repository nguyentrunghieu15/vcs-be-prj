package auth

import (
	"context"
	"fmt"
	"log"

	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/logger"
	auth "github.com/nguyentrunghieu15/vcs-common-prj/apu/auth"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/managedb"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type AuthServer struct {
	auth.AuthServiceServer
	UserRepo model.UserRepositoryDecorator
	Jwt      JwtService
	Bcrypt   BcryptService
	Logger   *logger.LoggerDecorator
}

type LogMessageAuth map[string]interface{}

func (s *AuthServer) Login(ctx context.Context, msg *auth.LoginMessage) (*auth.LoginResponse, error) {
	s.Logger.Log(
		logger.INFO,
		LogMessageAuth{
			"Action": "Ivoked Login",
			"User":   msg.GetEmail(),
		})
	// Check exist user by email
	user, err := s.UserRepo.FindOneByEmail(msg.Email)
	if err != nil {
		s.Logger.Log(logger.INFO,
			LogMessageAuth{
				"Action": "Ivoked Login",
				"User":   msg.GetEmail(),
				"Error":  err,
			})
		return nil, status.Errorf(codes.Internal, "Error : %v", err)
	}

	if user == nil {
		s.Logger.Log(
			logger.INFO,
			LogMessageAuth{
				"Action": "Ivoked Login",
				"User":   msg.GetEmail(),
				"Error":  "Not Found User",
			})
		return nil, status.Errorf(codes.NotFound, "Not found email")
	}

	// Check password
	if !s.Bcrypt.CheckPasswordHash(msg.Password, user.Password) {
		s.Logger.Log(
			logger.INFO,
			LogMessageAuth{
				"Action": "Ivoked Login",
				"User":   msg.GetEmail(),
				"Error":  "Wrong Password",
			})
		return nil, status.Errorf(codes.Unauthenticated, "Wrong pass word")
	}

	// Create access token
	accessToken, err := s.Jwt.GenerateAuthAccessToken(msg.GetEmail(), int64(user.ID), user.Roles)
	if err != nil {
		s.Logger.Log(
			logger.INFO,
			LogMessageAuth{
				"Action": "Ivoked Login",
				"User":   msg.GetEmail(),
				"Error":  err,
			})
		return nil, status.Errorf(codes.Internal, "Error : %v", err)
	}

	refreshToken, err := s.Jwt.GenerateAuthRefreshToken(msg.GetEmail(), int64(user.ID), user.Roles)
	if err != nil {
		s.Logger.Log(
			logger.INFO,
			LogMessageAuth{
				"Action": "Ivoked Login",
				"User":   msg.GetEmail(),
				"Error":  err,
			})
		return nil, status.Errorf(codes.Internal, "Error : %v", err)
	}

	return &auth.LoginResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3000,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthServer) RefreshToken(
	ctx context.Context,
	msg *auth.RefreshTokenMessage,
) (*auth.RefreshTokenResponse, error) {
	s.Logger.Log(
		logger.INFO,
		LogMessageAuth{
			"Action": "Ivoked Refresh token",
		})

	// Verify token
	if v, err := s.Jwt.VerifyToken(msg.GetRefreshToken()); v {
		if err != nil {
			s.Logger.Log(
				logger.INFO,
				LogMessageAuth{
					"Action": "Ivoked Refresh Token",
					"Token":  msg.GetRefreshToken(),
					"Error":  err})
		}
		return nil, nil
	}

	claims := s.Jwt.GetClaimsFromToken(msg.GetRefreshToken())
	// Check exist user by email
	user, err := s.UserRepo.FindOneByEmail(claims.Email)
	if err != nil {
		s.Logger.Log(
			logger.INFO,
			LogMessageAuth{"Action": "Ivoked Refresh Token",
				"Token": msg.GetRefreshToken(),
				"Error": err,
			},
		)
		return nil, err
	}

	if user == nil {
		s.Logger.Log(
			logger.INFO,
			LogMessageAuth{
				"Action": "Ivoked Refresh Token",
				"Token":  msg.GetRefreshToken(),
				"Error":  err,
			},
		)
		return nil, nil
	}
	// Create access token
	accessToken, err := s.Jwt.GenerateAuthAccessToken(claims.Email, int64(user.ID), user.Roles)
	if err != nil {
		s.Logger.Log(
			logger.INFO,
			LogMessageAuth{
				"Action": "Ivoked Refresh Token",
				"Token":  msg.GetRefreshToken(),
				"Error":  err,
			},
		)
		return nil, err
	}
	return &auth.RefreshTokenResponse{AccessToken: accessToken}, nil
}

func NewAuthServer() *AuthServer {
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
	var authServer AuthServer
	connPostgres, _ := postgres.(*gorm.DB)
	// connPostgres.Config.Logger = gormLogger.Default.LogMode(gormLogger.Silent)
	authServer.UserRepo = model.CreateUserRepository(connPostgres)
	authServer.Jwt.SecretKey = env.GetEnv("JWT_SECRETKEY").(string)
	newLogger := logger.NewLogger()
	newLogger.Config = logger.LoggerConfig{
		IsLogRotate:     true,
		PathToLog:       env.GetEnv("AUTH_LOG_PATH").(string),
		FileNameLogBase: env.GetEnv("AUTH_NAME_FILE_LOG").(string),
	}
	authServer.Logger = newLogger
	return &authServer
}
