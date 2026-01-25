package poolexp

import (
	"go.uber.org/fx"
)

var Module = fx.Module("adt/procexp",
	fx.Provide(
		fx.Annotate(newPgxDAO, fx.As(new(Repo))),
	),
)
