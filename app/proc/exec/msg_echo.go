package exec

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"smecalculus/rolevod/lib/id"
)

// Adapter
type handlerEcho struct {
	api API
	log *slog.Logger
}

func newHandlerEcho(a API, l *slog.Logger) *handlerEcho {
	return &handlerEcho{a, l}
}

func (h *handlerEcho) GetSnap(c echo.Context) error {
	var dto IdentMsg
	err := c.Bind(&dto)
	if err != nil {
		return err
	}
	idAttr := slog.Any("procID", dto.ProcID)
	id, err := id.ConvertFromString(dto.ProcID)
	if err != nil {
		h.log.Error("mapping failed", idAttr)
		return err
	}
	snap, err := h.api.Retrieve(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, MsgFromSnap(snap))
}

func (h *handlerEcho) PostCall(c echo.Context) error {
	var dto SpecMsg
	err := c.Bind(&dto)
	if err != nil {
		return err
	}
	idAttr := slog.Any("procID", dto.ProcID)
	spec, err := MsgToSpec(dto)
	if err != nil {
		h.log.Error("mapping failed", idAttr)
		return err
	}
	err = h.api.Run(spec)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}
