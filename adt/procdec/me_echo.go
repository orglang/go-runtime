package procdec

import (
	"log/slog"
	"net/http"
	"orglang/go-runtime/adt/uniqref"
	"reflect"

	"github.com/labstack/echo/v4"

	"github.com/orglang/go-sdk/adt/procdec"
)

// Server-side primary adapter
type echoController struct {
	api API
	log *slog.Logger
}

func newEchoController(a API, l *slog.Logger) *echoController {
	name := slog.String("name", reflect.TypeFor[echoController]().Name())
	return &echoController{a, l.With(name)}
}

func cfgEchoController(e *echo.Echo, h *echoController) error {
	e.POST("/api/v1/decs", h.PostSpec)
	e.GET("/api/v1/decs/:id", h.GetSnap)
	return nil
}

func (h *echoController) PostSpec(c echo.Context) error {
	var dto procdec.DecSpec
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
	spec, conversionErr := MsgToDecSpec(dto)
	if conversionErr != nil {
		h.log.Error("conversion failed", slog.Any("dto", dto))
		return conversionErr
	}
	ref, creationErr := h.api.Create(spec)
	if creationErr != nil {
		return creationErr
	}
	return c.JSON(http.StatusCreated, uniqref.MsgFromADT(ref))
}

func (h *echoController) GetSnap(c echo.Context) error {
	var dto procdec.DecRef
	bindingErr := c.Bind(&dto)
	if bindingErr != nil {
		h.log.Error("binding failed", slog.Any("dto", reflect.TypeOf(dto)))
		return bindingErr
	}
	ref, conversionErr := uniqref.MsgToADT(dto)
	if conversionErr != nil {
		h.log.Error("conversion failed", slog.Any("dto", dto))
		return conversionErr
	}
	snap, retrievalErr := h.api.RetrieveSnap(ref)
	if retrievalErr != nil {
		return retrievalErr
	}
	return c.JSON(http.StatusOK, MsgFromDecSnap(snap))
}
