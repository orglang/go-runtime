package lf

import (
	"go.uber.org/fx"
)

var Module = fx.Module("lib/lf",
	fx.Provide(
		newLoggerSlog,
	),
)
