package procexec

import (
	"log/slog"
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/procstep"
	"orglang/orglang/lib/lf"
)

// Server-side primary adapter
type echoHandler struct {
	api API
	log *slog.Logger
}

func newEchoHandler(a API, l *slog.Logger) *echoHandler {
	return &echoHandler{a, l}
}

func cfgEchoHandler(e *echo.Echo, h *echoHandler) error {
	e.GET("/api/v1/procs/:id", h.GetSnap)
	e.POST("/api/v1/procs/:id/execs", h.PostCall)
	e.POST("/api/v1/pools/:id/steps", h.PostOne)
	return nil
}

func (h *echoHandler) GetSnap(c echo.Context) error {
	var dto IdentME
	bindingErr := c.Bind(&dto)
	if bindingErr != nil {
		h.log.Error("binding failed", slog.Any("dto", dto))
		return bindingErr
	}
	id, conversionErr := identity.ConvertFromString(dto.ExecID)
	if conversionErr != nil {
		h.log.Error("conversion failed", slog.Any("dto", dto))
		return conversionErr
	}
	snap, retrievalErr := h.api.Retrieve(id)
	if retrievalErr != nil {
		return retrievalErr
	}
	return c.JSON(http.StatusOK, MsgFromExecSnap(snap))
}

func (h *echoHandler) PostCall(c echo.Context) error {
	var dto ExecSpecME
	bindingErr := c.Bind(&dto)
	if bindingErr != nil {
		h.log.Error("binding failed", slog.Any("dto", dto))
		return bindingErr
	}
	spec, conversionErr := MsgToExecSpec(dto)
	if conversionErr != nil {
		h.log.Error("conversion failed", slog.Any("dto", dto))
		return conversionErr
	}
	runningErr := h.api.Run(spec)
	if runningErr != nil {
		return runningErr
	}
	return c.NoContent(http.StatusOK)
}

func (h *echoHandler) PostOne(c echo.Context) error {
	var dto procstep.StepSpecME
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
	spec, conversionErr := procstep.MsgToStepSpec(dto)
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
