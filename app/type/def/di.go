//go:build !goverter

package def

import (
	"embed"
	"html/template"
	"log/slog"

	"github.com/Masterminds/sprig/v3"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"smecalculus/rolevod/lib/msg"
)

var Module = fx.Module("app/role",
	fx.Provide(
		fx.Annotate(newService, fx.As(new(API))),
		fx.Annotate(newRepoPgx, fx.As(new(Repo))),
	),
	fx.Provide(
		fx.Private,
		newHandlerEcho,
		newPresenterEcho,
		fx.Annotate(newRenderer, fx.As(new(msg.Renderer))),
	),
	fx.Invoke(
		cfgApiEcho,
		cfgSsrEcho,
	),
)

//go:embed *.html
var viewFs embed.FS

func newRenderer(l *slog.Logger) (*msg.RendererStdlib, error) {
	t, err := template.New("role").Funcs(sprig.FuncMap()).ParseFS(viewFs, "*.html")
	if err != nil {
		return nil, err
	}
	return msg.NewRendererStdlib(t, l), nil
}

func cfgApiEcho(e *echo.Echo, h *handlerEcho) error {
	e.POST("/api/v1/roles", h.PostOne)
	e.GET("/api/v1/roles/:id", h.GetOne)
	e.PATCH("/api/v1/roles/:id", h.PatchOne)
	return nil
}

func cfgSsrEcho(e *echo.Echo, p *presenterEcho) error {
	e.POST("/ssr/roles", p.PostOne)
	e.GET("/ssr/roles", p.GetMany)
	e.GET("/ssr/roles/:id", p.GetOne)
	return nil
}
