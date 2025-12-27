package web

import (
	"embed"
	"html/template"
	"log/slog"

	"github.com/Masterminds/sprig/v3"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"orglang/orglang/avt/msg"
)

var Module = fx.Module("app/web",
	fx.Provide(
		fx.Private,
		fx.Annotate(newRenderer, fx.As(new(msg.Renderer))),
		newHandlerEcho,
	),
	fx.Invoke(
		cfgEcho,
	),
)

//go:embed all:view
var webFs embed.FS

func newRenderer(l *slog.Logger) (*msg.RendererStdlib, error) {
	t, err := template.New("web").Funcs(sprig.FuncMap()).ParseFS(webFs, "*/*.html")
	if err != nil {
		return nil, err
	}
	return msg.NewRendererStdlib(t, l), nil
}

func cfgEcho(e *echo.Echo, h *handlerEcho) {
	e.GET("/", h.Home)
}
