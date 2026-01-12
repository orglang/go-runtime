//go:build !goverter

package procexec

import (
	"go.uber.org/fx"
)

var Module = fx.Module("adt/procexec",
	fx.Provide(
		fx.Annotate(newService, fx.As(new(API))),
	),
	fx.Provide(
		fx.Private,
		newEchoHandler,
		fx.Annotate(newPgxDAO, fx.As(new(Repo))),
	),
	fx.Invoke(
		cfgEchoHandler,
	),
)
