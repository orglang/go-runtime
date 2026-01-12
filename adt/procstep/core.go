package procstep

import (
	"fmt"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/procexp"
	"orglang/orglang/adt/revnum"
)

type StepSpec struct {
	ExecID identity.ADT
	ProcID identity.ADT
	ProcES procexp.ExpSpec
}

// aka Sem
type StepRec interface {
	step() identity.ADT
}

func ChnlID(r StepRec) identity.ADT { return r.step() }

type MsgRec struct {
	PoolID identity.ADT
	ExecID identity.ADT
	ChnlID identity.ADT
	ValER  procexp.ExpRec
	PoolRN revnum.ADT
	ProcRN revnum.ADT
}

func (r MsgRec) step() identity.ADT { return r.ChnlID }

type SvcRec struct {
	PoolID identity.ADT
	ExecID identity.ADT
	ChnlID identity.ADT
	ContER procexp.ExpRec
	PoolRN revnum.ADT
}

func (r SvcRec) step() identity.ADT { return r.ChnlID }

func ErrRecTypeUnexpected(got StepRec) error {
	return fmt.Errorf("step rec unexpected: %T", got)
}

func ErrRecTypeMismatch(got, want StepRec) error {
	return fmt.Errorf("step rec mismatch: want %T, got %T", want, got)
}
