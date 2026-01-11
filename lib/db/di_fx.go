package db

import (
	"go.uber.org/fx"
)

var Module = fx.Module("lib/db",
	fx.Provide(
		newPgxDriver,
		fx.Annotate(newOperator, fx.As(new(Operator))),
	),
	fx.Provide(
		fx.Private,
		newStorageCS,
	),
)
