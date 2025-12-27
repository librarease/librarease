package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/librarease/librarease/internal/config"

	"github.com/labstack/echo/v4"
)

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
			s.logger.ErrorContext(ctx, "AuthMiddleware: failed to get UID", slog.String("error", err.Error()))
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

// helper method to get uid from token
func (s *Server) getUID(c echo.Context) (string, error) {
	ctx := c.Request().Context()

	var (
		reqClientID = c.Request().Header.Get(config.HEADER_KEY_X_CLIENT_ID)
		reqUID      = c.Request().Header.Get(config.HEADER_KEY_X_UID)
		clientID    = os.Getenv(config.ENV_KEY_CLIENT_ID)
	)

	if reqClientID != "" &&
		reqUID != "" &&
		reqClientID == clientID {

		s.logger.InfoContext(ctx, "internal client request authenticated", slog.String("uid", reqUID))
		return reqUID, nil
	}

	// check for Authorization header
	var authH = c.Request().Header.Get("Authorization")

	if len(authH) >= len("Bearer ") {
		token := authH[len("Bearer "):]
		// FIXME: potential panic
		s.logger.DebugContext(ctx, "verifying bearer token", slog.String("token_prefix", token[:10]))

		return s.server.VerifyIDToken(c.Request().Context(), token)
	}

	// check for session cookie
	cname := os.Getenv(config.ENV_KEY_SESSION_COOKIE)
	authC, err := c.Request().Cookie(cname)
	if err != nil {
		s.logger.DebugContext(ctx, "session cookie not found", slog.String("cookie_name", cname), slog.String("err", err.Error()))
		return "", fmt.Errorf("cookie %s not found: %v", cname, err)
	}

	return s.server.VerifyIDToken(c.Request().Context(), authC.Value)
}
