package typedef

import (
	"log/slog"
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"

	"github.com/orglang/go-sdk/adt/typedef"

	"orglang/go-runtime/lib/lf"

	"orglang/go-runtime/adt/uniqref"
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
	e.POST("/api/v1/types", h.PostSpec)
	e.GET("/api/v1/types/:id", h.GetSnap)
	e.PATCH("/api/v1/types/:id", h.PatchOne)
	return nil
}

func (h *echoController) PostSpec(c echo.Context) error {
	var dto typedef.DefSpec
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
	spec, conversionErr := MsgToDefSpec(dto)
	if conversionErr != nil {
		h.log.Error("conversion failed", slog.Any("dto", dto))
		return conversionErr
	}
	snap, creationErr := h.api.Create(spec)
	if creationErr != nil {
		return creationErr
	}
	h.log.Log(ctx, lf.LevelTrace, "posting succeed", slog.Any("defRef", snap.DefRef))
	return c.JSON(http.StatusCreated, MsgFromDefSnap(snap))
}

func (h *echoController) GetSnap(c echo.Context) error {
	var dto typedef.DefRef
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
	ref, conversionErr := uniqref.MsgToADT(dto)
	if conversionErr != nil {
		h.log.Error("conversion failed", slog.Any("dto", dto))
		return conversionErr
	}
	snap, retrievalErr := h.api.RetrieveSnap(ref)
	if retrievalErr != nil {
		return retrievalErr
	}
	return c.JSON(http.StatusOK, MsgFromDefSnap(snap))
}

func (h *echoController) PatchOne(c echo.Context) error {
	var dto typedef.DefSnap
	bindingErr := c.Bind(&dto)
	if bindingErr != nil {
		h.log.Error("binding failed", slog.Any("dto", reflect.TypeOf(dto)))
		return bindingErr
	}
	ctx := c.Request().Context()
	h.log.Log(ctx, lf.LevelTrace, "patching started", slog.Any("dto", dto))
	validationErr := dto.Validate()
	if validationErr != nil {
		h.log.Error("validation failed", slog.Any("dto", dto))
		return validationErr
	}
	reqSnap, conversionErr := MsgToDefSnap(dto)
	if conversionErr != nil {
		h.log.Error("conversion failed", slog.Any("dto", dto))
		return conversionErr
	}
	resSnap, modificationErr := h.api.Modify(reqSnap)
	if modificationErr != nil {
		return modificationErr
	}
	h.log.Log(ctx, lf.LevelTrace, "patching succeed", slog.Any("defRef", resSnap.DefRef))
	return c.JSON(http.StatusOK, MsgFromDefSnap(resSnap))
}
