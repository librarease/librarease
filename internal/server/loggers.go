package server

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/librarease/librarease/internal/config"
)

func skipper(c echo.Context) bool {
	return slices.Contains([]string{
		"/api/health",
		"/favicon.ico",
	}, c.Request().URL.Path)
}

func NewEchoLogger(l *slog.Logger) echo.MiddlewareFunc {

	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:        true,
		LogURI:           true,
		LogError:         true,
		HandleError:      true, // forwards error to the global error handler, so it can decide appropriate status code
		LogRequestID:     true,
		LogRemoteIP:      true,
		LogHost:          true,
		LogMethod:        true,
		LogUserAgent:     true,
		LogLatency:       true,
		LogContentLength: true,
		LogResponseSize:  true,
		Skipper:          skipper,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			level := slog.LevelInfo
			msg := "REQUEST"
			if v.Error != nil {
				level = slog.LevelError
				msg = "REQUEST_ERROR"
			}

			attrs := []slog.Attr{
				slog.String("request_id", v.RequestID),
				slog.String("remote_ip", v.RemoteIP),
				slog.String("host", v.Host),
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.String("user_agent", v.UserAgent),
				slog.Int("status", v.Status),
				slog.Duration("latency", v.Latency),
				slog.String("latency_human", v.Latency.String()),
				slog.String("bytes_in", v.ContentLength),
				slog.Int64("bytes_out", v.ResponseSize),
				slog.String("protocol", c.Request().Proto),
				slog.String("uid", strings.Join(v.Headers[config.HEADER_KEY_X_UID], ",")),
				slog.String("client_id", strings.Join(v.Headers[config.HEADER_KEY_X_CLIENT_ID], ",")),
			}

			if v.Error != nil {
				attrs = append(attrs, slog.String("err", v.Error.Error()))
			}

			l.LogAttrs(c.Request().Context(), level, msg, attrs...)
			return nil
		},
	})
}
