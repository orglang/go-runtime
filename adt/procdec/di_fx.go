//go:build !goverter

package procdec

import (
	"go.uber.org/fx"

	"orglang/go-runtime/lib/te"
)

var Module = fx.Module("adt/procdec",
	fx.Provide(
		fx.Annotate(newService, fx.As(new(API))),
		fx.Annotate(newPgxDAO, fx.As(new(Repo))),
	),
	fx.Provide(
		fx.Private,
		newEchoHandler,
		newEchoPresenter,
		fx.Annotate(newRendererStdlib, fx.As(new(te.Renderer))),
	),
	fx.Invoke(
		cfgEchoHandler,
		cfgEchoPresenter,
	),
)
