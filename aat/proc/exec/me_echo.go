package exec

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"orglang/orglang/avt/id"
)

// Server-side primary adapter
type handlerEcho struct {
	api API
	log *slog.Logger
}

func newHandlerEcho(a API, l *slog.Logger) *handlerEcho {
	return &handlerEcho{a, l}
}

func cfgHandlerEcho(e *echo.Echo, h *handlerEcho) error {
	e.GET("/api/v1/procs/:id", h.GetSnap)
	e.POST("/api/v1/procs/:id/calls", h.PostCall)
	return nil
}

func (h *handlerEcho) GetSnap(c echo.Context) error {
	var dto IdentME
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
	var dto SpecME
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
