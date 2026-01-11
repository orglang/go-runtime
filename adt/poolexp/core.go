package poolexp

import (
	"orglang/orglang/adt/qualsym"
)

type ExpSpec interface {
	via()
}

type AcquireSpec struct {
	PoolQN qualsym.ADT
	// or
	ProcQN qualsym.ADT
	// or
	TypeQN qualsym.ADT
}

func (s AcquireSpec) via() {}

type AcceptSpec struct {
	CommPH qualsym.ADT
	ValPH  qualsym.ADT
}

func (s AcceptSpec) via() {}
