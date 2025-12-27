package server

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/librarease/librarease/internal/config"
	"go.opentelemetry.io/otel/trace"
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
		LogHeaders:       []string{config.HEADER_KEY_X_UID, config.HEADER_KEY_X_CLIENT_ID},
		Skipper:          skipper,

		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			level := slog.LevelInfo
			msg := "REQUEST"
			if v.Error != nil {
				level = slog.LevelError
				msg = "REQUEST_ERROR"
			}

			attrs := []slog.Attr{
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.Duration("latency", v.Latency),
				slog.Int64("bytes_out", v.ResponseSize),
				slog.String("remote_ip", v.RemoteIP),
			}

			// Add optional fields only if present
			if uid := strings.Join(v.Headers[config.HEADER_KEY_X_UID], ","); uid != "" {
				attrs = append(attrs, slog.String("uid", uid))
			}
			if clientID := strings.Join(v.Headers[config.HEADER_KEY_X_CLIENT_ID], ","); clientID != "" {
				attrs = append(attrs, slog.String("client_id", clientID))
			}

			// Extract trace context from the request context
			span := trace.SpanFromContext(c.Request().Context())
			if span.SpanContext().IsValid() {
				attrs = append(attrs,
					slog.String("trace_id", span.SpanContext().TraceID().String()),
					slog.String("span_id", span.SpanContext().SpanID().String()),
				)
			}

			if v.Error != nil {
				attrs = append(attrs, slog.String("err", v.Error.Error()))
			}

			l.LogAttrs(c.Request().Context(), level, msg, attrs...)
			return nil
		},
	})
}
