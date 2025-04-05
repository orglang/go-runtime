//go:build !goverter

package sig

import (
	"embed"
	"html/template"
	"log/slog"

	"github.com/Masterminds/sprig/v3"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"smecalculus/rolevod/lib/msg"
)

var Module = fx.Module("app/sig",
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
	t, err := template.New("sig").Funcs(sprig.FuncMap()).ParseFS(viewFs, "*.html")
	if err != nil {
		return nil, err
	}
	return msg.NewRendererStdlib(t, l), nil
}

func cfgApiEcho(e *echo.Echo, h *handlerEcho) error {
	e.POST("/api/v1/signatures", h.PostOne)
	e.GET("/api/v1/signatures/:id", h.GetOne)
	return nil
}

func cfgSsrEcho(e *echo.Echo, p *presenterEcho) error {
	e.POST("/ssr/signatures", p.PostOne)
	e.GET("/ssr/signatures", p.GetMany)
	e.GET("/ssr/signatures/:id", p.GetOne)
	return nil
}
