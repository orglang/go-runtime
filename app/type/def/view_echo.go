package def

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"smecalculus/rolevod/lib/core"
	"smecalculus/rolevod/lib/id"
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
	name := slog.String("name", "rolePresenterEcho")
	return &presenterEcho{a, r, l.With(name)}
}

func (p *presenterEcho) PostOne(c echo.Context) error {
	var dto TypeSpecView
	err := c.Bind(&dto)
	if err != nil {
		p.log.Error("dto binding failed")
		return err
	}
	ctx := c.Request().Context()
	p.log.Log(ctx, core.LevelTrace, "role posting started", slog.Any("dto", dto))
	err = dto.Validate()
	if err != nil {
		p.log.Error("dto validation failed")
		return err
	}
	ns, err := sym.ConvertFromString(dto.NS)
	if err != nil {
		p.log.Error("dto parsing failed")
		return err
	}
	snap, err := p.api.Create(TypeSpec{TypeSN: ns.New(dto.Name), TypeTS: OneSpec{}})
	if err != nil {
		p.log.Error("role creation failed")
		return err
	}
	html, err := p.ssr.Render("view-one", ViewFromTypeSnap(snap))
	if err != nil {
		p.log.Error("view rendering failed")
		return err
	}
	p.log.Log(ctx, core.LevelTrace, "role posting succeeded", slog.Any("ref", ConvertSnapToRef(snap)))
	return c.HTMLBlob(http.StatusOK, html)
}

func (p *presenterEcho) GetMany(c echo.Context) error {
	refs, err := p.api.RetreiveRefs()
	if err != nil {
		p.log.Error("refs retrieval failed")
		return err
	}
	html, err := p.ssr.Render("view-many", ViewFromTypeRefs(refs))
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
	id, err := id.ConvertFromString(dto.ID)
	if err != nil {
		p.log.Error("dto mapping failed")
		return err
	}
	snap, err := p.api.Retrieve(id)
	if err != nil {
		p.log.Error("root retrieval failed")
		return err
	}
	html, err := p.ssr.Render("view-one", ViewFromTypeSnap(snap))
	if err != nil {
		p.log.Error("view rendering failed")
		return err
	}
	p.log.Log(ctx, core.LevelTrace, "root getting succeeded", slog.Any("id", snap.TypeID))
	return c.HTMLBlob(http.StatusOK, html)
}
