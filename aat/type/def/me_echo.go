package def

import (
	"log/slog"
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"

	"orglang/orglang/lib/lf"

	"orglang/orglang/avt/id"
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
	e.POST("/api/v1/roles", h.PostOne)
	e.GET("/api/v1/roles/:id", h.GetOne)
	e.PATCH("/api/v1/roles/:id", h.PatchOne)
	return nil
}

func (h *handlerEcho) PostOne(c echo.Context) error {
	var dto TypeSpecME
	err := c.Bind(&dto)
	if err != nil {
		h.log.Error("dto binding failed")
		return err
	}
	ctx := c.Request().Context()
	h.log.Log(ctx, lf.LevelTrace, "role posting started", slog.Any("dto", dto))
	err = dto.Validate()
	if err != nil {
		h.log.Error("dto validation failed")
		return err
	}
	spec, err := MsgToTypeSpec(dto)
	if err != nil {
		h.log.Error("dto mapping failed")
		return err
	}
	snap, err := h.api.Create(spec)
	if err != nil {
		h.log.Error("role creation failed")
		return err
	}
	h.log.Log(ctx, lf.LevelTrace, "role posting succeeded", slog.Any("id", snap.TypeID))
	return c.JSON(http.StatusCreated, MsgFromTypeSnap(snap))
}

func (h *handlerEcho) GetOne(c echo.Context) error {
	var dto IdentME
	err := c.Bind(&dto)
	if err != nil {
		h.log.Error("dto binding failed")
		return err
	}
	err = dto.Validate()
	if err != nil {
		h.log.Error("dto validation failed")
		return err
	}
	id, err := id.ConvertFromString(dto.ID)
	if err != nil {
		h.log.Error("dto mapping failed")
		return err
	}
	snap, err := h.api.Retrieve(id)
	if err != nil {
		h.log.Error("root retrieval failed")
		return err
	}
	return c.JSON(http.StatusOK, MsgFromTypeSnap(snap))
}

func (h *handlerEcho) PatchOne(c echo.Context) error {
	var dto TypeSnapME
	err := c.Bind(&dto)
	if err != nil {
		h.log.Error("dto binding failed")
		return err
	}
	ctx := c.Request().Context()
	h.log.Log(ctx, lf.LevelTrace, "role patching started", slog.Any("dto", dto))
	err = dto.Validate()
	if err != nil {
		h.log.Error("dto validation failed")
		return err
	}
	reqSnap, err := MsgToTypeSnap(dto)
	if err != nil {
		h.log.Error("dto mapping failed")
		return err
	}
	resSnap, err := h.api.Modify(reqSnap)
	if err != nil {
		h.log.Error("role modification failed")
		return err
	}
	h.log.Log(ctx, lf.LevelTrace, "role patching succeeded", slog.Any("ref", ConvertSnapToRef(resSnap)))
	return c.JSON(http.StatusOK, MsgFromTypeSnap(resSnap))
}
