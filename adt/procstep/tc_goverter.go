package procstep

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
	DataToSemRecs   func([]StepRecDS) ([]StepRec, error)
	DataFromSemRecs func([]StepRec) ([]StepRecDS, error)
)
