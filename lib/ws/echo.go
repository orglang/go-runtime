package ws

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/fx"
)

func newServerEcho(pc exchangePC, l *slog.Logger, lc fx.Lifecycle) *echo.Echo {
	e := echo.New()
	log := l.With(slog.String("name", "serverEcho"))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod:   true,
		LogURI:      true,
		LogStatus:   true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error != nil {
				log.Error("request processing failed",
					slog.String("method", v.Method),
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("reason", v.Error.Error()),
				)
			}
			return nil
		},
	}))
	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				go e.Start(fmt.Sprintf(":%v", pc.Protocol.Http.Port))
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return e.Shutdown(ctx)
			},
		},
	)
	return e
}
