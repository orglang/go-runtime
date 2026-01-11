package poolstep

import (
	"orglang/orglang/adt/procexec"
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/adt/identity:Convert.*
// goverter:extend orglang/orglang/adt/procdef:Msg.*
var (
	MsgFromStepSpec func(StepSpec) StepSpecME
	MsgToStepSpec   func(StepSpecME) (StepSpec, error)
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/adt/identity:Convert.*
var (
	DataToLiab   func(liabDS) (procexec.Liab, error)
	DataFromLiab func(procexec.Liab) liabDS
	DataToEPs    func([]epDS) ([]procexec.EP, error)
)
