package auth

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthJwtClaims struct {
	Email string `json:"email"`
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

func (j *JwtService) GenerateAuthAccessToken(email string) (string, error) {
	expire_time, _ := strconv.Atoi(os.Getenv("EXPIRE_TIME"))
	authClaims := AuthJwtClaims{
		email,
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(expire_time))),
		},
	}
	return j.GenerateToken(authClaims)
}

func (j *JwtService) GenerateAuthRefreshToken(email string) (string, error) {
	authClaims := AuthJwtClaims{
		email,
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(9000))),
		},
	}
	return j.GenerateToken(authClaims)
}

func (j *JwtService) VerifyToken(token string) (bool, error) {
	result, err := jwt.ParseWithClaims(token, &AuthJwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.SecretKey), nil
	}, jwt.WithLeeway(5*time.Second))
	if err != nil {
		log.Printf("JWT Service: %v", err)
		return false, nil
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
