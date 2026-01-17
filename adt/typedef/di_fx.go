//go:build !goverter

package typedef

import (
	"go.uber.org/fx"

	"orglang/go-runtime/lib/te"
)

var Module = fx.Module("adt/typedef",
	fx.Provide(
		fx.Annotate(newService, fx.As(new(API))),
		fx.Annotate(newPgxDAO, fx.As(new(Repo))),
	),
	fx.Provide(
		fx.Private,
		newHandlerEcho,
		newEchoPresenter,
		fx.Annotate(newRendererStdlib, fx.As(new(te.Renderer))),
	),
	fx.Invoke(
		cfgHandlerEcho,
		cfgEchoPresenter,
	),
)
