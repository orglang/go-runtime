//go:build !goverter

package main

import (
	"go.uber.org/fx"

	"orglang/orglang/lib/lf"
	"orglang/orglang/lib/sd"
	"orglang/orglang/lib/ws"

	"orglang/orglang/adt/expalias"

	poolexec "orglang/orglang/adt/poolexec"
	"orglang/orglang/adt/procdecl"
	"orglang/orglang/adt/procdef"
	"orglang/orglang/adt/procexec"
	"orglang/orglang/adt/typedef"

	"orglang/orglang/app/web"
)

func main() {
	fx.New(
		// lib
		ws.Module,
		sd.Module,
		lf.Module,
		// aet
		expalias.Module,
		// aat
		procdef.Module,
		poolexec.Module,
		typedef.Module,
		procexec.Module,
		procdecl.Module,
		// app
		web.Module,
	).Run()
}
