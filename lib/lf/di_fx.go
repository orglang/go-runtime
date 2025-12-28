package lf

import (
	"go.uber.org/fx"
)

var Module = fx.Module("lib/lf", // logging framework
	fx.Provide(
		newLoggerSlog,
	),
)
