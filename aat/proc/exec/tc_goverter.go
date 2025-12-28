package exec

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/avt/id:Convert.*
// goverter:extend orglang/orglang/aat/proc/def:Msg.*
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
// goverter:extend orglang/orglang/avt/id:Convert.*
// goverter:extend orglang/orglang/avt/rn:Convert.*
// goverter:extend orglang/orglang/aat/proc/def:Data.*
// goverter:extend data.*
var (
	DataFromMod     func(Mod) (modDS, error)
	DataFromBnd     func(Bnd) bndDS
	DataToSemRecs   func([]SemRecDS) ([]SemRec, error)
	DataFromSemRecs func([]SemRec) ([]SemRecDS, error)
)
