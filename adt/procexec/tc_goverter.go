package procexec

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/adt/identity:Convert.*
// goverter:extend orglang/orglang/adt/proc/def:Msg.*
var (
	MsgFromSpec func(ProcSpec) SpecME
	MsgToSpec   func(SpecME) (ProcSpec, error)
	MsgToRef    func(RefME) (ProcRef, error)
	MsgFromRef  func(ProcRef) RefME
	MsgToSnap   func(SnapME) (ProcSnap, error)
	MsgFromSnap func(ProcSnap) SnapME
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/adt/identity:Convert.*
// goverter:extend orglang/orglang/adt/revnum:Convert.*
// goverter:extend orglang/orglang/adt/proc/def:Data.*
// goverter:extend data.*
var (
	DataFromMod     func(Mod) (modDS, error)
	DataFromBnd     func(Bnd) bndDS
	DataToSemRecs   func([]SemRecDS) ([]SemRec, error)
	DataFromSemRecs func([]SemRec) ([]SemRecDS, error)
)
