package server

import (
	"context"
	"fmt"
	"librarease/internal/config"
	"os"

	"github.com/labstack/echo/v4"
)

var (
	AppEnv  = os.Getenv(config.ENV_KEY_APP_ENV)
	isLocal = AppEnv == "local"
)

// func WithUserID() echo.MiddlewareFunc {
// 	return func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c echo.Context) error {
// 			if isLocal {
// 				userID := c.Request().Header.Get(config.HEADER_KEY_X_USER_ID)
// 				oc := c.Request().Context()
// 				nc := context.WithValue(oc, config.CTX_KEY_FB_UID, userID)
// 				c.SetRequest(c.Request().WithContext(nc))
// 				c.Set(string(config.CTX_KEY_FB_UID), userID)
// 				return next(c)
// 			}
// 			// TODO: Implement logic to get user ID from JWT token.
// 			return next(c)
// 		}
// 	}
// }

func (s *Server) getUID(c echo.Context) (string, error) {

	var (
		reqClientID = c.Request().Header.Get(config.HEADER_KEY_X_CLIENT_ID)
		reqUID      = c.Request().Header.Get(config.HEADER_KEY_X_UID)
		clientID    = os.Getenv(config.ENV_KEY_CLIENT_ID)
	)

	fmt.Printf("[AuthMiddleware] reqClientID: %s\n", reqClientID)
	fmt.Printf("[AuthMiddleware] reqUID: %s\n", reqUID)
	fmt.Printf("[AuthMiddleware] clientID: %s\n", clientID)

	if reqClientID != "" &&
		reqUID != "" &&
		reqClientID == clientID {

		fmt.Printf("[AuthMiddleware] internal client request: %s\n", reqUID)
		return reqUID, nil
	}

	var auth = c.Request().Header.Get("Authorization")

	if len(auth) < len("Bearer ") {
		return auth, c.JSON(401, map[string]string{"error": "Authorization header is required"})
	}

	token := auth[len("Bearer "):]
	fmt.Printf("[AuthMiddleware] token: %s...\n", token[:10])

	return s.server.VerifyIDToken(c.Request().Context(), token)
}

// AuthMiddleware check authorization header and verify the token
// using injected server.VerifyIDToken method, transforms request
// to have Firebase UID value in downstream context.
func (s *Server) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		var (
			ctx = c.Request().Context()
		)

		uid, err := s.getUID(c)

		if err != nil {
			fmt.Println("[AuthMiddleware] error: ", err)
			return c.JSON(401, map[string]string{
				"error":   err.Error(),
				"message": "Invalid token",
			})
		}

		au, err := s.server.GetAuthUserByUID(ctx, uid)
		if err != nil {
			return c.JSON(401, map[string]string{
				"error":   err.Error(),
				"message": "User not found",
			})
		}

		ctx = context.WithValue(ctx, config.CTX_KEY_USER_ID, au.UserID)
		ctx = context.WithValue(ctx, config.CTX_KEY_USER_ROLE, au.GlobalRole)

		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}
