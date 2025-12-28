//go:build !goverter

package dec

import (
	"go.uber.org/fx"

	"orglang/orglang/lib/te"
)

var Module = fx.Module("proc/dec",
	fx.Provide(
		fx.Annotate(newService, fx.As(new(API))),
		fx.Annotate(newDaoPgx, fx.As(new(Repo))),
	),
	fx.Provide(
		fx.Private,
		newHandlerEcho,
		newPresenterEcho,
		fx.Annotate(newRendererStdlib, fx.As(new(te.Renderer))),
	),
	fx.Invoke(
		cfgHandlerEcho,
		cfgPresenterEcho,
	),
)
