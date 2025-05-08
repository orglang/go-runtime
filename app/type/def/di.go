//go:build !goverter

package def

import (
	"go.uber.org/fx"
)

var Module = fx.Module("internal/state",
	fx.Provide(
		fx.Annotate(newRepoPgx, fx.As(new(TermRepo))),
	),
)
