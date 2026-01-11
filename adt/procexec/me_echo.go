package procexec

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"orglang/orglang/adt/identity"
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
	e.POST("/api/v1/procs/:id/calls", h.PostCall)
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
