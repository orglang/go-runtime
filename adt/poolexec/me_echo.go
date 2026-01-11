package poolexec

import (
	"log/slog"
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"

	"orglang/orglang/adt/identity"
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
	e.POST("/api/v1/pools", h.PostOne)
	e.GET("/api/v1/pools/:id", h.GetOne)
	e.POST("/api/v1/pools/:id/procs", h.PostProc)
	return nil
}

func (h *echoHandler) PostOne(c echo.Context) error {
	var dto ExecSpecME
	bindingErr := c.Bind(&dto)
	if bindingErr != nil {
		h.log.Error("binding failed", slog.Any("dto", reflect.TypeOf(dto)))
		return bindingErr
	}
	validationErr := dto.Validate()
	if validationErr != nil {
		h.log.Error("validation failed", slog.Any("dto", dto))
		return validationErr
	}
	spec, conversionErr := MsgToExecSpec(dto)
	if conversionErr != nil {
		h.log.Error("conversion failed", slog.Any("dto", dto))
		return conversionErr
	}
	ref, creationErr := h.api.Run(spec)
	if creationErr != nil {
		return creationErr
	}
	return c.JSON(http.StatusCreated, MsgFromExecRef(ref))
}

func (h *echoHandler) GetOne(c echo.Context) error {
	var dto IdentME
	bindingErr := c.Bind(&dto)
	if bindingErr != nil {
		return bindingErr
	}
	id, conversionErr := identity.ConvertFromString(dto.PoolID)
	if conversionErr != nil {
		return conversionErr
	}
	snap, retrievalErr := h.api.Retrieve(id)
	if retrievalErr != nil {
		return retrievalErr
	}
	return c.JSON(http.StatusOK, MsgFromExecSnap(snap))
}

func (h *echoHandler) PostProc(c echo.Context) error {
	var dto ExecSpecME
	bindingErr := c.Bind(&dto)
	if bindingErr != nil {
		h.log.Error("binding failed", slog.Any("dto", reflect.TypeOf(dto)))
		return bindingErr
	}
	validationErr := dto.Validate()
	if validationErr != nil {
		h.log.Error("validation failed", slog.Any("dto", dto))
		return validationErr
	}
	spec, conversionErr := MsgToExecSpec(dto)
	if conversionErr != nil {
		h.log.Error("conversion failed", slog.Any("dto", dto))
		return conversionErr
	}
	ref, creationErr := h.api.Run(spec)
	if creationErr != nil {
		return creationErr
	}
	return c.JSON(http.StatusCreated, MsgFromExecRef(ref))
}
