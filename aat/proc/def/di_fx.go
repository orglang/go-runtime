//go:build !goverter

package def

import (
	"go.uber.org/fx"
)

var Module = fx.Module("proc/def",
	fx.Provide(
		fx.Annotate(newRepoPgx, fx.As(new(Repo))),
	),
)
