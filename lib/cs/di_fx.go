package cs

import (
	"go.uber.org/fx"
)

var Module = fx.Module("lib/cs", // configuration source
	fx.Provide(
		fx.Annotate(newKeeperViper, fx.As(new(Keeper))),
	),
)
