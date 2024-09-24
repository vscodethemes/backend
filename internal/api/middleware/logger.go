package middleware

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func Logger(logger *slog.Logger) echo.MiddlewareFunc {
	return echomiddleware.RequestLoggerWithConfig(echomiddleware.RequestLoggerConfig{
		LogStatus:    true,
		LogMethod:    true,
		LogURI:       true,
		LogRoutePath: true,
		LogRemoteIP:  true,
		LogRequestID: true,
		LogError:     true,
		HandleError:  true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			msg := fmt.Sprintf("%s %s - %d", v.Method, v.URI, v.Status)

			args := []slog.Attr{
				slog.Int("status", v.Status),
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.String("route_path", v.RoutePath),
				slog.String("remote_ip", v.RemoteIP),
				slog.String("request_id", v.RequestID),
			}

			level := slog.LevelInfo
			if v.Error != nil {
				level = slog.LevelError
				args = append(args, slog.String("err", v.Error.Error()))
			}

			logger.LogAttrs(context.Background(), level, msg, args...)

			return nil
		},
	})
}
