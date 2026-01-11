package typedef

import (
	"log/slog"
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"

	"orglang/orglang/lib/lf"

	"orglang/orglang/adt/identity"
)

// Server-side primary adapter
type handlerEcho struct {
	api API
	log *slog.Logger
}

func newHandlerEcho(a API, l *slog.Logger) *handlerEcho {
	name := slog.String("name", reflect.TypeFor[handlerEcho]().Name())
	return &handlerEcho{a, l.With(name)}
}

func cfgHandlerEcho(e *echo.Echo, h *handlerEcho) error {
	e.POST("/api/v1/types", h.PostOne)
	e.GET("/api/v1/types/:id", h.GetOne)
	e.PATCH("/api/v1/types/:id", h.PatchOne)
	return nil
}

func (h *handlerEcho) PostOne(c echo.Context) error {
	var dto DefSpecME
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
	h.log.Log(ctx, lf.LevelTrace, "posting succeed", slog.Any("id", snap.DefID))
	return c.JSON(http.StatusCreated, MsgFromDefSnap(snap))
}

func (h *handlerEcho) GetOne(c echo.Context) error {
	var dto IdentME
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
	id, conversionErr := identity.ConvertFromString(dto.DefID)
	if conversionErr != nil {
		h.log.Error("conversion failed", slog.Any("dto", dto))
		return conversionErr
	}
	snap, retrievalErr := h.api.Retrieve(id)
	if retrievalErr != nil {
		return retrievalErr
	}
	return c.JSON(http.StatusOK, MsgFromDefSnap(snap))
}

func (h *handlerEcho) PatchOne(c echo.Context) error {
	var dto DefSnapME
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
	h.log.Log(ctx, lf.LevelTrace, "patching succeed", slog.Any("ref", ConvertSnapToRef(resSnap)))
	return c.JSON(http.StatusOK, MsgFromDefSnap(resSnap))
}
