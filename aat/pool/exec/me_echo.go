package exec

import (
	"log/slog"
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"

	"orglang/orglang/avt/core"
	"orglang/orglang/avt/id"
	"orglang/orglang/avt/msg"

	procexec "orglang/orglang/aat/proc/exec"
)

// Adapter
type handlerEcho struct {
	api API
	ssr msg.Renderer
	log *slog.Logger
}

func newHandlerEcho(a API, r msg.Renderer, l *slog.Logger) *handlerEcho {
	name := slog.String("name", "poolHandlerEcho")
	return &handlerEcho{a, r, l.With(name)}
}

func (h *handlerEcho) PostOne(c echo.Context) error {
	var dto PoolSpecMsg
	err := c.Bind(&dto)
	if err != nil {
		h.log.Error("binding failed", slog.Any("struct", reflect.TypeOf(dto)))
		return err
	}
	qnAttr := slog.Any("sigQN", dto.SigQN)
	err = dto.Validate()
	if err != nil {
		h.log.Error("validation failed", qnAttr)
		return err
	}
	spec, err := MsgToPoolSpec(dto)
	if err != nil {
		h.log.Error("mapping failed", qnAttr)
		return err
	}
	ref, err := h.api.Create(spec)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, MsgFromPoolRef(ref))
}

func (h *handlerEcho) GetOne(c echo.Context) error {
	var dto IdentMsg
	err := c.Bind(&dto)
	if err != nil {
		return err
	}
	id, err := id.ConvertFromString(dto.PoolID)
	if err != nil {
		return err
	}
	snap, err := h.api.Retrieve(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, MsgFromPoolSnap(snap))
}

func (h *handlerEcho) PostProc(c echo.Context) error {
	var dto PoolSpecMsg
	err := c.Bind(&dto)
	if err != nil {
		h.log.Error("binding failed", slog.Any("struct", reflect.TypeOf(dto)))
		return err
	}
	qnAttr := slog.Any("sigQN", dto.SigQN)
	err = dto.Validate()
	if err != nil {
		h.log.Error("validation failed", qnAttr)
		return err
	}
	spec, err := MsgToPoolSpec(dto)
	if err != nil {
		h.log.Error("mapping failed", qnAttr)
		return err
	}
	ref, err := h.api.Create(spec)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, MsgFromPoolRef(ref))
}

// Adapter
type stepHandlerEcho struct {
	api API
	ssr msg.Renderer
	log *slog.Logger
}

func newStepHandlerEcho(a API, r msg.Renderer, l *slog.Logger) *stepHandlerEcho {
	name := slog.String("name", "stepHandlerEcho")
	return &stepHandlerEcho{a, r, l.With(name)}
}

func (h *stepHandlerEcho) PostOne(c echo.Context) error {
	var dto procexec.SpecMsg
	err := c.Bind(&dto)
	if err != nil {
		h.log.Error("binding failed")
		return err
	}
	ctx := c.Request().Context()
	h.log.Log(ctx, core.LevelTrace, "posting started", slog.Any("dto", dto))
	err = dto.Validate()
	if err != nil {
		h.log.Error("validation failed", slog.Any("dto", dto))
		return err
	}
	spec, err := procexec.MsgToSpec(dto)
	if err != nil {
		h.log.Error("mapping failed", slog.Any("dto", dto))
		return err
	}
	ref, err := h.api.Spawn(spec)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, procexec.MsgFromRef(ref))
}
