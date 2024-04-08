package auth

import (
	"context"
	"log"
	"net/http"
	"os"

	auth "github.com/nguyentrunghieu15/vcs-common-prj/apu/auth"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/managedb"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type AuthServer struct {
	auth.AuthServiceServer
	UserRepo UserRepositoryDecorator
	Jwt      JwtService
	Bcrypt   BcryptService
}

func (s *AuthServer) Login(ctx context.Context, msg *auth.LoginMessage) (*auth.LoginResponse, error) {
	// Check exist user by email
	user, err := s.UserRepo.FindOneByEmail(msg.Email)
	if err != nil {
		GrpcStatus{http.StatusInternalServerError}.AddStatus(&ctx)
		return nil, status.Errorf(codes.Internal, "Error : %v", err)
	}

	if user == nil {
		GrpcStatus{http.StatusNotFound}.AddStatus(&ctx)
		return nil, status.Errorf(codes.NotFound, "Not found email")
	}

	// Check password
	if !s.Bcrypt.CheckPasswordHash(msg.Password, user.Password) {
		GrpcStatus{http.StatusUnauthorized}.AddStatus(&ctx)
		return nil, status.Errorf(codes.Unauthenticated, "Wrong pass word")
	}

	// Create access token
	accessToken, err := s.Jwt.GenerateAuthAccessToken(msg.Email)
	if err != nil {
		GrpcStatus{http.StatusInternalServerError}.AddStatus(&ctx)
		return nil, status.Errorf(codes.Internal, "Error : %v", err)
	}

	refreshToken, err := s.Jwt.GenerateAuthRefreshToken(msg.Email)
	if err != nil {
		GrpcStatus{http.StatusInternalServerError}.AddStatus(&ctx)
		return nil, status.Errorf(codes.Internal, "Error : %v", err)
	}
	GrpcStatus{http.StatusAccepted}.AddStatus(&ctx)

	return &auth.LoginResponse{AccessToken: accessToken, TokenType: "Jwt", ExpiresIn: 3000, RefreshToken: refreshToken}, nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, msg *auth.RefreshTokenMessage) (*auth.RefreshTokenResponse, error) {
	// Verify token
	if !s.Jwt.VerifyToken(msg.GetRefreshToken()) {
		GrpcStatus{http.StatusUnauthorized}.AddStatus(&ctx)
		return nil, nil
	}

	claims := s.Jwt.GetClaimsFromToken(msg.GetRefreshToken())
	// Check exist user by email
	user, err := s.UserRepo.FindOneByEmail(claims.Email)
	if err != nil {
		GrpcStatus{http.StatusInternalServerError}.AddStatus(&ctx)
		return nil, err
	}

	if user == nil {
		GrpcStatus{http.StatusNotFound}.AddStatus(&ctx)
		return nil, nil
	}
	// Create access token
	accessToken, err := s.Jwt.GenerateAuthAccessToken(claims.Email)
	if err != nil {
		GrpcStatus{http.StatusInternalServerError}.AddStatus(&ctx)
		return nil, err
	}
	return &auth.RefreshTokenResponse{AccessToken: accessToken}, nil
}

func CreateAuthServer() *AuthServer {
	dsnPostgres := "host=localhost user=hiro password=1 dbname=vcs_msm_db port=5432 sslmode=disable"
	postgres, err := managedb.GetConnection(managedb.Connection{Context: &managedb.PostgreContext{}, Dsn: dsnPostgres})
	if err != nil {
		log.Fatalf("AuthService : Can't connect to PostgresSQL Database :%v", err)
	}
	log.Println("Connected database", postgres)
	var authServer AuthServer
	connPostgres, _ := postgres.(*gorm.DB)
	authServer.UserRepo.UserRepository = *model.CreateUserRepository(connPostgres)
	authServer.Jwt.SecretKey = os.Getenv("JWT_SECRET_KEY")
	return &authServer
}
