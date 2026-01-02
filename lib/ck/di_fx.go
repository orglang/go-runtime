package ck

import (
	"go.uber.org/fx"
)

var Module = fx.Module("lib/ck", // configuration keeper
	fx.Provide(
		fx.Annotate(newKeeperViper, fx.As(new(Loader))),
	),
)
