//go:build !goverter

package syndec

import (
	"go.uber.org/fx"
)

var Module = fx.Module("adt/syndec",
	fx.Provide(
		fx.Annotate(newPgxDAO, fx.As(new(Repo))),
	),
)
