package typeexp

import (
	"go.uber.org/fx"
)

var Module = fx.Module("adt/typeexp",
	fx.Provide(
		fx.Annotate(newPgxDAO, fx.As(new(Repo))),
	),
)
