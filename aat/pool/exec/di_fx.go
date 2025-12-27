//go:build !goverter

package exec

import (
	"embed"
	"html/template"
	"log/slog"

	"github.com/Masterminds/sprig/v3"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"orglang/orglang/avt/msg"
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
		fx.Annotate(newRenderer, fx.As(new(msg.Renderer))),
	),
	fx.Invoke(
		cfgEcho,
		cfgStepEcho,
	),
)

//go:embed *.html
var viewsFs embed.FS

func newRenderer(l *slog.Logger) (*msg.RendererStdlib, error) {
	t, err := template.New("pool").Funcs(sprig.FuncMap()).ParseFS(viewsFs, "*.html")
	if err != nil {
		return nil, err
	}
	return msg.NewRendererStdlib(t, l), nil
}

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
