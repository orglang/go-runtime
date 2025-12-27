package msg

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/fx"

	"orglang/orglang/avt/core"
)

var Module = fx.Module("lib/msg",
	fx.Provide(
		newEcho,
	),
	fx.Provide(
		fx.Private,
		newCfg,
	),
)

func newCfg(k core.Keeper) (*props, error) {
	props := &props{}
	err := k.Load("messaging", props)
	if err != nil {
		return nil, err
	}
	return props, nil
}

func newEcho(p *props, l *slog.Logger, lc fx.Lifecycle) *echo.Echo {
	e := echo.New()
	log := l.With(slog.String("name", "echo.Echo"))
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
				go e.Start(fmt.Sprintf(":%v", p.Protocol.Http.Port))
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return e.Shutdown(ctx)
			},
		},
	)
	return e
}
