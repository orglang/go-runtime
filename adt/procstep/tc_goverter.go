package procstep

import "github.com/orglang/go-sdk/adt/procstep"

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
// goverter:extend orglang/go-runtime/adt/procexp:Msg.*
var (
	MsgFromStepSpec func(StepSpec) procstep.StepSpec
	MsgToStepSpec   func(procstep.StepSpec) (StepSpec, error)
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
// goverter:extend data.*
var (
	DataToStepRecs   func([]StepRecDS) ([]StepRec, error)
	DataFromStepRecs func([]StepRec) ([]StepRecDS, error)
)
