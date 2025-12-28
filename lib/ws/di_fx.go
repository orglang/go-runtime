package ws

import (
	"go.uber.org/fx"
)

var Module = fx.Module("lib/ws", // web server
	fx.Provide(
		newServerEcho,
	),
	fx.Provide(
		fx.Private,
		newExchangeConfig,
	),
)
