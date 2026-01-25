package main

import (
	"go.uber.org/fx"

	"orglang/go-runtime/lib/db"
	"orglang/go-runtime/lib/lf"
	"orglang/go-runtime/lib/ws"

	"orglang/go-runtime/adt/poolexec"
	"orglang/go-runtime/adt/procdec"
	"orglang/go-runtime/adt/procdef"
	"orglang/go-runtime/adt/procexec"
	"orglang/go-runtime/adt/syndec"
	"orglang/go-runtime/adt/typedef"

	"orglang/go-runtime/app/web"
)

func main() {
	fx.New(
		// lib
		db.Module,
		lf.Module,
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
