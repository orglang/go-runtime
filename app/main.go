//go:build !goverter

package main

import (
	"go.uber.org/fx"

	"orglang/orglang/avt/core"
	"orglang/orglang/avt/data"
	"orglang/orglang/avt/msg"

	"orglang/orglang/aet/alias"

	poolexec "orglang/orglang/aat/pool/exec"
	procdec "orglang/orglang/aat/proc/dec"
	procdef "orglang/orglang/aat/proc/def"
	procexec "orglang/orglang/aat/proc/exec"
	typedef "orglang/orglang/aat/type/def"

	"orglang/orglang/app/web"
)

func main() {
	fx.New(
		// lib
		core.Module,
		data.Module,
		msg.Module,
		alias.Module,
		// app
		procdef.Module,
		poolexec.Module,
		typedef.Module,
		procexec.Module,
		procdec.Module,
		web.Module,
	).Run()
}
