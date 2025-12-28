package sd

import (
	"go.uber.org/fx"

	"orglang/orglang/lib/cs"
)

var Module = fx.Module("lib/sd", // storage driver
	fx.Provide(
		newPgx,
		fx.Annotate(newOperator, fx.As(new(Operator))),
	),
	fx.Provide(
		fx.Private,
		newCfg,
	),
)

func newCfg(k cs.Keeper) (*props, error) {
	props := &props{}
	err := k.Load("storage", props)
	if err != nil {
		return nil, err
	}
	return props, nil
}
