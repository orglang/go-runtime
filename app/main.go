//go:build !goverter

package main

import (
	"go.uber.org/fx"

	"smecalculus/rolevod/lib/core"
	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/msg"

	"smecalculus/rolevod/internal/alias"

	pooldef "smecalculus/rolevod/app/pool/def"
	procdec "smecalculus/rolevod/app/proc/dec"
	procdef "smecalculus/rolevod/app/proc/def"
	proceval "smecalculus/rolevod/app/proc/eval"
	typedec "smecalculus/rolevod/app/type/dec"
	typedef "smecalculus/rolevod/app/type/def"

	"smecalculus/rolevod/app/web"
)

func main() {
	fx.New(
		// lib
		core.Module,
		data.Module,
		msg.Module,
		alias.Module,
		// app
		typedef.Module,
		procdef.Module,
		pooldef.Module,
		typedec.Module,
		proceval.Module,
		procdec.Module,
		web.Module,
	).Run()
}
