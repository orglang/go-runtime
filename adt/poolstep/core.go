package poolstep

import (
	"orglang/go-runtime/adt/poolexp"
	"orglang/go-runtime/adt/uniqref"
	"orglang/go-runtime/adt/uniqsym"
)

type StepSpec struct {
	ExecRef uniqref.ADT
	ProcQN  uniqsym.ADT
	ProcES  poolexp.ExpSpec
}
