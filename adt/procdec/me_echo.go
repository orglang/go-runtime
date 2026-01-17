package procdec

import (
	"log/slog"
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"

	"github.com/orglang/go-sdk/adt/procdec"

	"orglang/go-runtime/adt/identity"
)

// Server-side primary adapter
type echoHandler struct {
	api API
	log *slog.Logger
}

func newEchoHandler(a API, l *slog.Logger) *echoHandler {
	name := slog.String("name", reflect.TypeFor[echoHandler]().Name())
	return &echoHandler{a, l.With(name)}
}

func cfgEchoHandler(e *echo.Echo, h *echoHandler) error {
	e.POST("/api/v1/decs", h.PostOne)
	e.GET("/api/v1/decs/:id", h.GetOne)
	return nil
}

func (h *echoHandler) PostOne(c echo.Context) error {
	var dto procdec.DecSpecME
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
	snap, creationErr := h.api.Create(spec)
	if creationErr != nil {
		return creationErr
	}
	return c.JSON(http.StatusCreated, MsgFromDecSnap(snap))
}

func (h *echoHandler) GetOne(c echo.Context) error {
	var dto procdec.IdentME
	bindingErr := c.Bind(&dto)
	if bindingErr != nil {
		h.log.Error("binding failed", slog.Any("dto", reflect.TypeOf(dto)))
		return bindingErr
	}
	id, conversionErr := identity.ConvertFromString(dto.DecID)
	if conversionErr != nil {
		h.log.Error("conversion failed", slog.Any("dto", dto))
		return conversionErr
	}
	snap, retrievalErr := h.api.RetrieveSnap(id)
	if retrievalErr != nil {
		return retrievalErr
	}
	return c.JSON(http.StatusOK, MsgFromDecSnap(snap))
}
