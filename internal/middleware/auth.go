package middleware

import (
	"net/http"
	"search-job/internal/pkg/jwt"
	"strings"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "missing authorization header",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid authorization header format",
			})
		}

		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid or expired token",
			})
		}

		c.Set("user_id", claims.UserID)
		return next(c)
	}
}

func GetUserID(c echo.Context) int64 {
	userID, ok := c.Get("user_id").(int64)
	if !ok {
		return 0
	}
	return userID
}
