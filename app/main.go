//go:build !goverter

package main

import (
	"go.uber.org/fx"

	"smecalculus/rolevod/lib/core"
	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/msg"

	"smecalculus/rolevod/internal/alias"
	"smecalculus/rolevod/internal/chnl"
	"smecalculus/rolevod/internal/state"
	"smecalculus/rolevod/internal/step"

	"smecalculus/rolevod/app/pool"
	"smecalculus/rolevod/app/role"
	"smecalculus/rolevod/app/sig"
	"smecalculus/rolevod/app/web"
)

func main() {
	fx.New(
		// lib
		core.Module,
		data.Module,
		msg.Module,
		// internal
		alias.Module,
		chnl.Module,
		state.Module,
		step.Module,
		// app
		pool.Module,
		role.Module,
		sig.Module,
		web.Module,
	).Run()
}
