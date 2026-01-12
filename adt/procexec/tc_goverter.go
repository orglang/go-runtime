package procexec

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/adt/identity:Convert.*
// goverter:extend orglang/orglang/adt/procdef:Msg.*
var (
	MsgFromExecSpec func(ExecSpec) ExecSpecME
	MsgToExecSpec   func(ExecSpecME) (ExecSpec, error)
	MsgToExecRef    func(ExecRefME) (ExecRef, error)
	MsgFromExecRef  func(ExecRef) ExecRefME
	MsgToExecSnap   func(ExecSnapME) (ExecSnap, error)
	MsgFromExecSnap func(ExecSnap) ExecSnapME
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/adt/identity:Convert.*
// goverter:extend orglang/orglang/adt/revnum:Convert.*
// goverter:extend orglang/orglang/adt/procdef:Data.*
// goverter:extend data.*
var (
	DataFromMod  func(Mod) (modDS, error)
	DataFromBnd  func(Bnd) bindDS
	DataToLiab   func(liabDS) (Liab, error)
	DataFromLiab func(Liab) liabDS
	DataToEPs    func([]epDS) ([]EP, error)
)
