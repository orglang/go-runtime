package dec

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"orglang/orglang/avt/id"
	"orglang/orglang/avt/sym"
	"orglang/orglang/lib/lf"
	"orglang/orglang/lib/te"
)

// Adapter
type presenterEcho struct {
	api API
	ssr te.Renderer
	log *slog.Logger
}

func newPresenterEcho(a API, r te.Renderer, l *slog.Logger) *presenterEcho {
	name := slog.String("name", "sigPresenterEcho")
	return &presenterEcho{a, r, l.With(name)}
}

func cfgPresenterEcho(e *echo.Echo, p *presenterEcho) error {
	e.POST("/ssr/signatures", p.PostOne)
	e.GET("/ssr/signatures", p.GetMany)
	e.GET("/ssr/signatures/:id", p.GetOne)
	return nil
}

func (p *presenterEcho) PostOne(c echo.Context) error {
	var dto SigSpecVP
	err := c.Bind(&dto)
	if err != nil {
		p.log.Error("dto binding failed")
		return err
	}
	ctx := c.Request().Context()
	p.log.Log(ctx, lf.LevelTrace, "root posting started", slog.Any("dto", dto))
	err = dto.Validate()
	if err != nil {
		p.log.Error("dto validation failed")
		return err
	}
	ns, err := sym.ConvertFromString(dto.SigNS)
	if err != nil {
		p.log.Error("dto parsing failed")
		return err
	}
	ref, err := p.api.Incept(ns.New(dto.SigSN))
	if err != nil {
		p.log.Error("root creation failed")
		return err
	}
	html, err := p.ssr.Render("view-one", ViewFromSigRef(ref))
	if err != nil {
		p.log.Error("view rendering failed")
		return err
	}
	p.log.Log(ctx, lf.LevelTrace, "root posting succeeded", slog.Any("ref", ref))
	return c.HTMLBlob(http.StatusOK, html)
}

func (p *presenterEcho) GetMany(c echo.Context) error {
	refs, err := p.api.RetreiveRefs()
	if err != nil {
		p.log.Error("refs retrieval failed")
		return err
	}
	html, err := p.ssr.Render("view-many", ViewFromSigRefs(refs))
	if err != nil {
		p.log.Error("view rendering failed")
		return err
	}
	return c.HTMLBlob(http.StatusOK, html)
}

func (p *presenterEcho) GetOne(c echo.Context) error {
	var dto IdentME
	err := c.Bind(&dto)
	if err != nil {
		p.log.Error("dto binding failed")
		return err
	}
	ctx := c.Request().Context()
	p.log.Log(ctx, lf.LevelTrace, "root getting started", slog.Any("dto", dto))
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
	html, err := p.ssr.Render("view-one", ViewFromSigSnap(snap))
	if err != nil {
		p.log.Error("view rendering failed")
		return err
	}
	p.log.Log(ctx, lf.LevelTrace, "root getting succeeded", slog.Any("id", snap.DecID))
	return c.HTMLBlob(http.StatusOK, html)
}
