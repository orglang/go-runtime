//go:build !goverter

package exec

import (
	"go.uber.org/fx"
)

var Module = fx.Module("proc/exec",
	fx.Provide(
		fx.Annotate(newService, fx.As(new(API))),
	),
	fx.Provide(
		fx.Private,
		newHandlerEcho,
		fx.Annotate(newDaoPgx, fx.As(new(repo))),
	),
	fx.Invoke(
		cfgHandlerEcho,
	),
)
