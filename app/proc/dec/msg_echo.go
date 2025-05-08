package dec

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/msg"
)

// Adapter
type handlerEcho struct {
	api API
	ssr msg.Renderer
	log *slog.Logger
}

func newHandlerEcho(a API, r msg.Renderer, l *slog.Logger) *handlerEcho {
	name := slog.String("name", "sigHandlerEcho")
	return &handlerEcho{a, r, l.With(name)}
}

func (h *handlerEcho) PostOne(c echo.Context) error {
	var dto SigSpecMsg
	err := c.Bind(&dto)
	if err != nil {
		h.log.Error("dto binding failed", slog.Any("reason", err))
		return err
	}
	err = dto.Validate()
	if err != nil {
		h.log.Error("dto validation failed", slog.Any("reason", err), slog.Any("dto", dto))
		return err
	}
	spec, err := MsgToSigSpec(dto)
	if err != nil {
		h.log.Error("dto conversion failed", slog.Any("reason", err), slog.Any("dto", dto))
		return err
	}
	snap, err := h.api.Create(spec)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, MsgFromSigSnap(snap))
}

func (h *handlerEcho) GetOne(c echo.Context) error {
	var dto IdentMsg
	err := c.Bind(&dto)
	if err != nil {
		return err
	}
	id, err := id.ConvertFromString(dto.SigID)
	if err != nil {
		return err
	}
	snap, err := h.api.Retrieve(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, MsgFromSigSnap(snap))
}
