package procdec

import (
	"log/slog"
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"

	"orglang/go-runtime/lib/lf"
	"orglang/go-runtime/lib/te"

	"orglang/go-runtime/adt/uniqref"
	"orglang/go-runtime/adt/uniqsym"
)

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
	e.POST("/ssr/decs", p.PostSpec)
	e.GET("/ssr/decs", p.GetRefs)
	e.GET("/ssr/decs/:id", p.GetSnap)
	return nil
}

func (p *echoPresenter) PostSpec(c echo.Context) error {
	var dto DecSpecVP
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
	qn, conversionErr := uniqsym.ConvertFromString(dto.ProcQN)
	if conversionErr != nil {
		p.log.Error("conversion failed", slog.Any("dto", dto))
		return conversionErr
	}
	ref, inceptionErr := p.api.Incept(qn)
	if inceptionErr != nil {
		return inceptionErr
	}
	html, renderingErr := p.ssr.Render("view-one", uniqref.MsgFromADT(ref))
	if renderingErr != nil {
		p.log.Error("rendering failed", slog.Any("ref", ref))
		return renderingErr
	}
	p.log.Log(ctx, lf.LevelTrace, "posting succeed", slog.Any("ref", ref))
	return c.HTMLBlob(http.StatusOK, html)
}

func (p *echoPresenter) GetRefs(c echo.Context) error {
	refs, retrievalErr := p.api.RetreiveRefs()
	if retrievalErr != nil {
		return retrievalErr
	}
	html, renderingErr := p.ssr.Render("view-many", uniqref.MsgFromADTs(refs))
	if renderingErr != nil {
		p.log.Error("rendering failed", slog.Any("refs", refs))
		return renderingErr
	}
	return c.HTMLBlob(http.StatusOK, html)
}

func (p *echoPresenter) GetSnap(c echo.Context) error {
	var dto DecRefVP
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
	ref, conversionErr := uniqref.MsgToADT(dto)
	if conversionErr != nil {
		p.log.Error("conversion failed", slog.Any("dto", dto))
		return conversionErr
	}
	snap, retrievalErr := p.api.RetrieveSnap(ref)
	if retrievalErr != nil {
		return retrievalErr
	}
	html, renderingErr := p.ssr.Render("view-one", ViewFromDecSnap(snap))
	if renderingErr != nil {
		p.log.Error("rendering failed", slog.Any("snap", snap))
		return renderingErr
	}
	p.log.Log(ctx, lf.LevelTrace, "getting succeed", slog.Any("decRef", snap.DecRef))
	return c.HTMLBlob(http.StatusOK, html)
}
