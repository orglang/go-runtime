//go:build !goverter

package exec

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

var Module = fx.Module("aat/pool",
	fx.Provide(
		fx.Annotate(newService, fx.As(new(API))),
	),
	fx.Provide(
		fx.Private,
		newHandlerEcho,
		newStepHandlerEcho,
		fx.Annotate(newDaoPgx, fx.As(new(Repo))),
	),
	fx.Invoke(
		cfgEcho,
		cfgStepEcho,
	),
)

func cfgEcho(e *echo.Echo, h *handlerEcho) error {
	e.POST("/api/v1/pools", h.PostOne)
	e.GET("/api/v1/pools/:id", h.GetOne)
	e.POST("/api/v1/pools/:id/procs", h.PostProc)
	return nil
}

func cfgStepEcho(e *echo.Echo, h *stepHandlerEcho) error {
	e.POST("/api/v1/pools/:id/steps", h.PostOne)
	return nil
}
