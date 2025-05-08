//go:build !goverter

package def

import (
	"go.uber.org/fx"
)

var Module = fx.Module("internal/step",
	fx.Provide(
		fx.Annotate(newRepoPgx, fx.As(new(SemRepo))),
	),
)
