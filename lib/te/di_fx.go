package te

import (
	"go.uber.org/fx"
)

var Module = fx.Module("lib/te", // template engine
	fx.Provide(),
)
