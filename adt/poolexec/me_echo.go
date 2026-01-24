package poolexec

import (
	"log/slog"
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"

	"github.com/orglang/go-sdk/adt/poolexec"

	"orglang/go-runtime/lib/te"
)

// Server-side primary adapter
type echoController struct {
	api API
	ssr te.Renderer
	log *slog.Logger
}

func newEchoController(a API, r te.Renderer, l *slog.Logger) *echoController {
	name := slog.String("name", reflect.TypeFor[echoController]().Name())
	return &echoController{a, r, l.With(name)}
}

func cfgEchoController(e *echo.Echo, h *echoController) error {
	e.POST("/api/v1/pools", h.PostOne)
	e.GET("/api/v1/pools/:id", h.GetOne)
	e.POST("/api/v1/pools/:id/procs", h.PostProc)
	return nil
}

func (h *echoController) PostOne(c echo.Context) error {
	var dto poolexec.ExecSpec
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

func (h *echoController) GetOne(c echo.Context) error {
	var dto poolexec.ExecRef
	bindingErr := c.Bind(&dto)
	if bindingErr != nil {
		return bindingErr
	}
	ref, conversionErr := MsgToExecRef(dto)
	if conversionErr != nil {
		return conversionErr
	}
	snap, retrievalErr := h.api.RetrieveSnap(ref)
	if retrievalErr != nil {
		return retrievalErr
	}
	return c.JSON(http.StatusOK, MsgFromExecSnap(snap))
}

func (h *echoController) PostProc(c echo.Context) error {
	var dto poolexec.ExecSpec
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
