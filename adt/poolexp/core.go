package poolexp

import (
	"orglang/orglang/adt/qualsym"
)

type ExpSpec interface {
	via()
}

type HireSpec struct {
	ProcQN qualsym.ADT
}

func (s HireSpec) via() {}

type FireSpec struct {
	ProcQN qualsym.ADT
}

func (s FireSpec) via() {}

type ApplySpec struct {
	ProcQN qualsym.ADT
}

func (s ApplySpec) via() {}

type QuitSpec struct {
	ProcQN qualsym.ADT
}

func (s QuitSpec) via() {}

type AcquireSpec struct {
	PoolQN qualsym.ADT
	BindPH qualsym.ADT
}

func (s AcquireSpec) via() {}

type ReleaseSpec struct {
}

func (s ReleaseSpec) via() {}

type AcceptSpec struct {
	PoolQN qualsym.ADT
	ValPH  qualsym.ADT
}

func (s AcceptSpec) via() {}

type DetachSpec struct {
	PoolQN qualsym.ADT
	ValPH  qualsym.ADT
}

func (s DetachSpec) via() {}
