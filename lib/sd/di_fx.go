package sd

import (
	"go.uber.org/fx"
)

var Module = fx.Module("lib/sd", // storage driver
	fx.Provide(
		newDriverPgx,
		fx.Annotate(newOperator, fx.As(new(Operator))),
	),
	fx.Provide(
		fx.Private,
		newStoragePC,
	),
)
