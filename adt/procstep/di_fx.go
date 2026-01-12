//go:build !goverter

package procstep

import (
	"go.uber.org/fx"
)

var Module = fx.Module("adt/procstep",
	fx.Provide(
		fx.Annotate(newPgxDAO, fx.As(new(Repo))),
	),
)
