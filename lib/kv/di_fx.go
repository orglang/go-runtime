package kv

import (
	"go.uber.org/fx"
)

var Module = fx.Module("lib/kv",
	fx.Provide(
		fx.Annotate(newViperDriver, fx.As(new(Loader))),
	),
)
