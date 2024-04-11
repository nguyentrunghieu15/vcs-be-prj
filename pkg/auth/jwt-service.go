package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
)

type AuthJwtClaims struct {
	Email string `json:"email"`
	Id    int64  `json:"id"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

type JwtService struct {
	SecretKey string
}

func (j *JwtService) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(j.SecretKey))
	return ss, err
}

func (j *JwtService) GenerateAuthAccessToken(email string, id int64, role model.UserRole) (string, error) {
	expire_time := env.GetEnv("JWT_EXPIRE_TIME").(int)
	fmt.Println(email, id, role)
	authClaims := AuthJwtClaims{
		email,
		id,
		string(role),
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(time.Second * time.Duration(expire_time))),
		},
	}
	return j.GenerateToken(authClaims)
}

func (j *JwtService) GenerateAuthRefreshToken(email string, id int64, role model.UserRole) (string, error) {
	authClaims := AuthJwtClaims{
		email,
		id,
		string(role),
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(time.Second * time.Duration(9000))),
		},
	}
	return j.GenerateToken(authClaims)
}

func (j *JwtService) VerifyToken(token string) (bool, error) {
	result, err := jwt.ParseWithClaims(token, &AuthJwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.SecretKey), nil
	}, jwt.WithLeeway(5*time.Second))
	if err != nil {
		return false, fmt.Errorf("JWT Service: %v", err)
	} else if _, ok := result.Claims.(*AuthJwtClaims); ok {
		return true, nil
	} else {
		return false, fmt.Errorf("JWT Service: Unknown claims type, cannot proceed")

	}
}

func (j *JwtService) GetClaimsFromToken(token string) *AuthJwtClaims {
	result, _ := jwt.ParseWithClaims(token, &AuthJwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.SecretKey), nil
	}, jwt.WithLeeway(5*time.Second))
	if claims, ok := result.Claims.(*AuthJwtClaims); ok {
		return claims
	}
	return nil
}
