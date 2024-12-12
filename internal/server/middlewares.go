package server

import (
	"librarease/internal/config"
	"os"

	"github.com/labstack/echo/v4"
)

var (
	AppEnv  = os.Getenv(config.ENV_KEY_APP_ENV)
	isLocal = AppEnv == "local"
)

func WithUserID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if isLocal {
				userID := c.Request().Header.Get(config.HEADER_KEY_X_USER_ID)
				c.Set(config.HEADER_KEY_X_USER_ID, userID)
				return next(c)
			}
			// TODO: Implement logic to get user ID from JWT token.
			return next(c)
		}
	}
}
