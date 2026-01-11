//go:build !goverter

package main

import (
	"go.uber.org/fx"

	"orglang/orglang/lib/db"
	"orglang/orglang/lib/lf"
	"orglang/orglang/lib/rc"
	"orglang/orglang/lib/ws"

	"orglang/orglang/adt/poolexec"
	"orglang/orglang/adt/procdec"
	"orglang/orglang/adt/procdef"
	"orglang/orglang/adt/procexec"
	"orglang/orglang/adt/syndec"
	"orglang/orglang/adt/typedef"

	"orglang/orglang/app/web"
)

func main() {
	fx.New(
		// lib
		db.Module,
		lf.Module,
		rc.Module,
		ws.Module,
		// adt
		syndec.Module,
		poolexec.Module,
		typedef.Module,
		procdef.Module,
		procdec.Module,
		procexec.Module,
		// app
		web.Module,
	).Run()
}
