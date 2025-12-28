package wf

import (
	"go.uber.org/fx"

	"orglang/orglang/lib/cs"
)

var Module = fx.Module("lib/wf", // web framework
	fx.Provide(
		newEcho,
	),
	fx.Provide(
		fx.Private,
		newCfg,
	),
)

func newCfg(k cs.Keeper) (*props, error) {
	props := &props{}
	err := k.Load("messaging", props)
	if err != nil {
		return nil, err
	}
	return props, nil
}
