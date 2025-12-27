//go:build !goverter

package exec

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

var Module = fx.Module("app/proc",
	fx.Provide(
		fx.Annotate(newService, fx.As(new(API))),
	),
	fx.Provide(
		fx.Private,
		newHandlerEcho,
		fx.Annotate(newRepoPgx, fx.As(new(repo))),
	),
	fx.Invoke(
		cfgEcho,
	),
)

func cfgEcho(e *echo.Echo, h *handlerEcho) error {
	e.GET("/api/v1/procs/:id", h.GetSnap)
	e.POST("/api/v1/procs/:id/calls", h.PostCall)
	return nil
}
