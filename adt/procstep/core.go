package procstep

import (
	"fmt"

	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/procexp"
	"orglang/go-runtime/adt/revnum"
	"orglang/go-runtime/adt/uniqref"
)

type StepSpec struct {
	ExecRef uniqref.ADT
	ProcES  procexp.ExpSpec
}

// aka Sem
type StepRec interface {
	step() identity.ADT
}

func ChnlID(rec StepRec) identity.ADT { return rec.step() }

type MsgRec struct {
	PoolRN  revnum.ADT
	ExecRef uniqref.ADT
	ChnlID  identity.ADT
	ValER   procexp.ExpRec
}

func (r MsgRec) step() identity.ADT { return r.ChnlID }

type SvcRec struct {
	PoolRN  revnum.ADT
	ExecRef uniqref.ADT
	ChnlID  identity.ADT
	ContER  procexp.ExpRec
}

func (r SvcRec) step() identity.ADT { return r.ChnlID }

func ErrRecTypeUnexpected(got StepRec) error {
	return fmt.Errorf("step rec unexpected: %T", got)
}

func ErrRecTypeMismatch(got, want StepRec) error {
	return fmt.Errorf("step rec mismatch: want %T, got %T", want, got)
}
