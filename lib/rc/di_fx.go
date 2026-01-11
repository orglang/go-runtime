package rc

import (
	"go.uber.org/fx"
)

var Module = fx.Module("lib/rc",
	fx.Provide(
		newRestyClient,
	),
)
