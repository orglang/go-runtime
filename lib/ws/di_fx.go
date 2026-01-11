package ws

import (
	"go.uber.org/fx"
)

var Module = fx.Module("lib/ws",
	fx.Provide(
		newEchoServer,
	),
	fx.Provide(
		fx.Private,
		newExchangeCS,
	),
)
