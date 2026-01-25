package procdef

import (
	"go.uber.org/fx"
)

var Module = fx.Module("adt/procdef",
	fx.Provide(
		fx.Annotate(newPgxDAO, fx.As(new(Repo))),
	),
)
