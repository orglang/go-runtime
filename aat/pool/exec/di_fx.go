//go:build !goverter

package exec

import (
	"go.uber.org/fx"
)

var Module = fx.Module("aat/pool",
	fx.Provide(
		fx.Annotate(newService, fx.As(new(API))),
	),
	fx.Provide(
		fx.Private,
		newHandlerEcho,
		newStepHandlerEcho,
		fx.Annotate(newDaoPgx, fx.As(new(Repo))),
	),
	fx.Invoke(
		cfgHandlerEcho,
		cfgStepHandlerEcho,
	),
)
