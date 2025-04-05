package sig

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"smecalculus/rolevod/lib/core"
	id "smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/msg"
	"smecalculus/rolevod/lib/sym"
)

// Adapter
type presenterEcho struct {
	api API
	ssr msg.Renderer
	log *slog.Logger
}

func newPresenterEcho(a API, r msg.Renderer, l *slog.Logger) *presenterEcho {
	name := slog.String("name", "sigPresenterEcho")
	return &presenterEcho{a, r, l.With(name)}
}

func (p *presenterEcho) PostOne(c echo.Context) error {
	var dto SpecView
	err := c.Bind(&dto)
	if err != nil {
		p.log.Error("dto binding failed")
		return err
	}
	ctx := c.Request().Context()
	p.log.Log(ctx, core.LevelTrace, "root posting started", slog.Any("dto", dto))
	err = dto.Validate()
	if err != nil {
		p.log.Error("dto validation failed")
		return err
	}
	fqn := sym.CovertFromString(dto.NS).New(dto.Name)
	ref, err := p.api.Incept(fqn)
	if err != nil {
		p.log.Error("root creation failed")
		return err
	}
	html, err := p.ssr.Render("view-one", ViewFromRef(ref))
	if err != nil {
		p.log.Error("view rendering failed")
		return err
	}
	p.log.Log(ctx, core.LevelTrace, "root posting succeeded", slog.Any("ref", ref))
	return c.HTMLBlob(http.StatusOK, html)
}

func (p *presenterEcho) GetMany(c echo.Context) error {
	refs, err := p.api.RetreiveRefs()
	if err != nil {
		p.log.Error("refs retrieval failed")
		return err
	}
	html, err := p.ssr.Render("view-many", ViewFromRefs(refs))
	if err != nil {
		p.log.Error("view rendering failed")
		return err
	}
	return c.HTMLBlob(http.StatusOK, html)
}

func (p *presenterEcho) GetOne(c echo.Context) error {
	var dto IdentMsg
	err := c.Bind(&dto)
	if err != nil {
		p.log.Error("dto binding failed")
		return err
	}
	ctx := c.Request().Context()
	p.log.Log(ctx, core.LevelTrace, "root getting started", slog.Any("dto", dto))
	err = dto.Validate()
	if err != nil {
		p.log.Error("dto validation failed")
		return err
	}
	id, err := id.ConvertFromString(dto.SigID)
	if err != nil {
		p.log.Error("dto mapping failed")
		return err
	}
	snap, err := p.api.Retrieve(id)
	if err != nil {
		p.log.Error("snap retrieval failed")
		return err
	}
	html, err := p.ssr.Render("view-one", ViewFromRoot(snap))
	if err != nil {
		p.log.Error("view rendering failed")
		return err
	}
	p.log.Log(ctx, core.LevelTrace, "root getting succeeded", slog.Any("id", snap.SigID))
	return c.HTMLBlob(http.StatusOK, html)
}
