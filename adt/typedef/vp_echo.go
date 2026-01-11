package typedef

import (
	"log/slog"
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"

	"orglang/orglang/lib/lf"
	"orglang/orglang/lib/te"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/qualsym"
	"orglang/orglang/adt/typeexp"
)

// Adapter
type echoPresenter struct {
	api API
	ssr te.Renderer
	log *slog.Logger
}

func newEchoPresenter(a API, r te.Renderer, l *slog.Logger) *echoPresenter {
	name := slog.String("name", reflect.TypeFor[echoPresenter]().Name())
	return &echoPresenter{a, r, l.With(name)}
}

func cfgEchoPresenter(e *echo.Echo, p *echoPresenter) error {
	e.POST("/ssr/types", p.PostOne)
	e.GET("/ssr/types", p.GetMany)
	e.GET("/ssr/types/:id", p.GetOne)
	return nil
}

func (p *echoPresenter) PostOne(c echo.Context) error {
	var dto DefSpecVP
	bindingErr := c.Bind(&dto)
	if bindingErr != nil {
		p.log.Error("binding failed", slog.Any("dto", reflect.TypeOf(dto)))
		return bindingErr
	}
	ctx := c.Request().Context()
	p.log.Log(ctx, lf.LevelTrace, "posting started", slog.Any("dto", dto))
	validationErr := dto.Validate()
	if validationErr != nil {
		p.log.Error("validation failed", slog.Any("dto", dto))
		return validationErr
	}
	ns, conversionErr := qualsym.ConvertFromString(dto.TypeNS)
	if conversionErr != nil {
		p.log.Error("conversion failed", slog.Any("dto", dto))
		return conversionErr
	}
	snap, creationErr := p.api.Create(DefSpec{TypeQN: ns.New(dto.TypeSN), TypeES: typeexp.OneSpec{}})
	if creationErr != nil {
		return creationErr
	}
	html, renderingErr := p.ssr.Render("view-one", ViewFromDefSnap(snap))
	if renderingErr != nil {
		p.log.Error("rendering failed", slog.Any("snap", snap))
		return renderingErr
	}
	p.log.Log(ctx, lf.LevelTrace, "posting succeed", slog.Any("ref", ConvertSnapToRef(snap)))
	return c.HTMLBlob(http.StatusOK, html)
}

func (p *echoPresenter) GetMany(c echo.Context) error {
	refs, retrievalErr := p.api.RetreiveRefs()
	if retrievalErr != nil {
		return retrievalErr
	}
	html, renderingErr := p.ssr.Render("view-many", ViewFromDefRefs(refs))
	if renderingErr != nil {
		p.log.Error("rendering failed", slog.Any("refs", refs))
		return renderingErr
	}
	return c.HTMLBlob(http.StatusOK, html)
}

func (p *echoPresenter) GetOne(c echo.Context) error {
	var dto IdentME
	bindingErr := c.Bind(&dto)
	if bindingErr != nil {
		p.log.Error("binding failed", slog.Any("dto", reflect.TypeOf(dto)))
		return bindingErr
	}
	ctx := c.Request().Context()
	p.log.Log(ctx, lf.LevelTrace, "getting started", slog.Any("dto", dto))
	validationErr := dto.Validate()
	if validationErr != nil {
		p.log.Error("validation failed", slog.Any("dto", dto))
		return validationErr
	}
	id, conversionErr := identity.ConvertFromString(dto.DefID)
	if conversionErr != nil {
		p.log.Error("conversion failed", slog.Any("dto", dto))
		return conversionErr
	}
	snap, retrievalErr := p.api.Retrieve(id)
	if retrievalErr != nil {
		return retrievalErr
	}
	html, renderingErr := p.ssr.Render("view-one", ViewFromDefSnap(snap))
	if renderingErr != nil {
		p.log.Error("rendering failed", slog.Any("snap", snap))
		return renderingErr
	}
	p.log.Log(ctx, lf.LevelTrace, "getting succeed", slog.Any("ref", ConvertSnapToRef(snap)))
	return c.HTMLBlob(http.StatusOK, html)
}
