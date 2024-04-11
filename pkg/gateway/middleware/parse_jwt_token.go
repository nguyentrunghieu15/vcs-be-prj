package gateway_middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/auth"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
)

func UserParseJWTTokenMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authorization := c.Request().Header.Get("Authorization")
			splitedAuthorization := strings.Split(authorization, " ")
			token := splitedAuthorization[1]
			tokenType := splitedAuthorization[0]

			result, _ := jwt.ParseWithClaims(token, &auth.AuthJwtClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(env.GetEnv("JWT_SECRETKEY").(string)), nil
			}, jwt.WithLeeway(5*time.Second))
			claims, _ := result.Claims.(*auth.AuthJwtClaims)

			c.Request().Header.Set("Grpc-Metadata-email", claims.Email)
			c.Request().Header.Set("Grpc-Metadata-id", fmt.Sprintf("%v", claims.Id))
			c.Request().Header.Set("Grpc-Metadata-role", claims.Role)
			c.Request().Header.Set("Grpc-Metadata-token-type", tokenType)

			return next(c)
		}
	}
}
