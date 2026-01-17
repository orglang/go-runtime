package web

import (
	"log/slog"
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"

	"orglang/go-runtime/adt/typedef"
	"orglang/go-runtime/lib/te"
)

// Adapter
type handlerEcho struct {
	api typedef.API
	ssr te.Renderer
	log *slog.Logger
}

func newHandlerEcho(a typedef.API, r te.Renderer, l *slog.Logger) *handlerEcho {
	name := slog.String("name", reflect.TypeFor[handlerEcho]().Name())
	return &handlerEcho{a, r, l.With(name)}
}

func cfgHandlerEcho(e *echo.Echo, h *handlerEcho) {
	e.GET("/", h.Home)
}

func (h *handlerEcho) Home(c echo.Context) error {
	refs, err := h.api.RetreiveRefs()
	if err != nil {
		return err
	}
	html, err := h.ssr.Render("home.html", typedef.MsgFromDefRefs(refs))
	if err != nil {
		return err
	}
	return c.HTMLBlob(http.StatusOK, html)
}
