package def

import (
	typedef "smecalculus/rolevod/app/type/def"
)

type TermSpec interface {
	poolDef()
}

type CloseSpec struct{}

func (s CloseSpec) poolDef() {}

type WaitSpec struct{}

func (s WaitSpec) poolDef() {}

type SendSpec struct {
	TypeTS typedef.TermSpec
}

func (s SendSpec) poolDef() {}

type RecvSpec struct {
	TypeTS typedef.TermSpec
}

func (s RecvSpec) poolDef() {}

type LabSpec struct{}

func (s LabSpec) poolDef() {}

type CaseSpec struct{}

func (s CaseSpec) poolDef() {}

type CallSpec struct{}

func (s CallSpec) poolDef() {}

type SpawnSpec struct{}

func (s SpawnSpec) poolDef() {}

type FwdSpec struct{}

func (s FwdSpec) poolDef() {}
