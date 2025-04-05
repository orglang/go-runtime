//go:build !goverter

package main

import (
	"go.uber.org/fx"

	"smecalculus/rolevod/lib/core"
	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/msg"

	"smecalculus/rolevod/internal/alias"
	"smecalculus/rolevod/internal/state"
	"smecalculus/rolevod/internal/step"

	pool "smecalculus/rolevod/app/pool/root"
	procroot "smecalculus/rolevod/app/proc/root"
	procsig "smecalculus/rolevod/app/proc/sig"
	role "smecalculus/rolevod/app/role/root"
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
		state.Module,
		step.Module,
		// app
		pool.Module,
		role.Module,
		procroot.Module,
		procsig.Module,
		web.Module,
	).Run()
}
