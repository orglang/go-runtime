package e2e

import (
	"go.uber.org/fx"
)

var Module = fx.Module("lib/e2e",
	fx.Provide(
		fx.Private,
		newPoolDecAPI,
		newPoolExecAPI,
		newPoolDecAPI,
		newProcExecAPI,
		newTypeDefAPI,
	),
)
