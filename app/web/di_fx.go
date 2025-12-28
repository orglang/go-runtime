package web

import (
	"go.uber.org/fx"

	"orglang/orglang/lib/te"
)

var Module = fx.Module("app/web",
	fx.Provide(
		fx.Private,
		fx.Annotate(newRendererStdlib, fx.As(new(te.Renderer))),
		newHandlerEcho,
	),
	fx.Invoke(
		cfgHandlerEcho,
	),
)
