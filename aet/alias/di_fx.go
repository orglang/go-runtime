//go:build !goverter

package alias

import (
	"go.uber.org/fx"
)

var Module = fx.Module("internal/alias",
	fx.Provide(
		fx.Annotate(newRepoPgx, fx.As(new(Repo))),
	),
)
