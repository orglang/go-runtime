package poolstep

import (
	"log/slog"
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"

	"orglang/orglang/lib/lf"
	"orglang/orglang/lib/te"
)

// Server-side primary adapter
type echoHandler struct {
	api API
	ssr te.Renderer
	log *slog.Logger
}

func newEchoHandler(a API, r te.Renderer, l *slog.Logger) *echoHandler {
	name := slog.String("name", reflect.TypeFor[echoHandler]().Name())
	return &echoHandler{a, r, l.With(name)}
}

func cfgEchoHandler(e *echo.Echo, h *echoHandler) error {
	e.POST("/api/v1/pools/:id/steps", h.PostOne)
	return nil
}

func (h *echoHandler) PostOne(c echo.Context) error {
	var dto StepSpecME
	bindingErr := c.Bind(&dto)
	if bindingErr != nil {
		h.log.Error("binding failed", slog.Any("dto", reflect.TypeOf(dto)))
		return bindingErr
	}
	ctx := c.Request().Context()
	h.log.Log(ctx, lf.LevelTrace, "posting started", slog.Any("dto", dto))
	validationErr := dto.Validate()
	if validationErr != nil {
		h.log.Error("validation failed", slog.Any("dto", dto))
		return validationErr
	}
	spec, conversionErr := MsgToStepSpec(dto)
	if conversionErr != nil {
		h.log.Error("conversion failed", slog.Any("dto", dto))
		return conversionErr
	}
	takingErr := h.api.Take(spec)
	if takingErr != nil {
		return takingErr
	}
	return c.NoContent(http.StatusOK)
}
