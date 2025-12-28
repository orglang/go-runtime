package dec

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/avt/id:Convert.*
var (
	ConvertSnapToRef func(ProcSnap) ProcRef
	ConvertRecToRef  func(ProcRec) ProcRef
	ConvertRecToSnap func(ProcRec) ProcSnap
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/avt/sym:Convert.*
// goverter:extend orglang/orglang/aat/type/def:Msg.*
var (
	MsgToBndSpec    func(BndSpecME) (ChnlSpec, error)
	MsgFromBndSpec  func(ChnlSpec) BndSpecME
	MsgToSigSpec    func(SigSpecME) (ProcSpec, error)
	MsgFromSigSpec  func(ProcSpec) SigSpecME
	MsgToSigRef     func(SigRefME) (ProcRef, error)
	MsgFromSigRef   func(ProcRef) SigRefME
	MsgToSigSnap    func(SigSnapME) (ProcSnap, error)
	MsgFromSigSnap  func(ProcSnap) SigSnapME
	MsgFromSigSnaps func([]ProcSnap) []SigSnapME
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/avt/id:Convert.*
// goverter:extend orglang/orglang/avt/rn:Convert.*
// goverter:extend orglang/orglang/aat/type/def:Msg.*
// goverter:extend Msg.*
var (
	ViewFromSigRef  func(ProcRef) SigRefVP
	ViewToSigRef    func(SigRefVP) (ProcRef, error)
	ViewFromSigRefs func([]ProcRef) []SigRefVP
	ViewToSigRefs   func([]SigRefVP) ([]ProcRef, error)
	ViewFromSigSnap func(ProcSnap) SigSnapVP
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/avt/id:Convert.*
// goverter:extend orglang/orglang/aat/type/def:Data.*
var (
	DataToSigRef     func(sigRefDS) (ProcRef, error)
	DataFromSigRef   func(ProcRef) sigRefDS
	DataToSigRefs    func([]sigRefDS) ([]ProcRef, error)
	DataFromSigRefs  func([]ProcRef) []sigRefDS
	DataToSigRec     func(sigRecDS) (ProcRec, error)
	DataFromSigRec   func(ProcRec) (sigRecDS, error)
	DataToSigRecs    func([]sigRecDS) ([]ProcRec, error)
	DataFromSigRecs  func([]ProcRec) ([]sigRecDS, error)
	DataToSigSnap    func(sigSnapDS) (ProcSnap, error)
	DataFromSigSnap  func(ProcSnap) (sigSnapDS, error)
	DataToSigSnaps   func([]sigSnapDS) ([]ProcSnap, error)
	DataFromSigSnaps func([]ProcSnap) ([]sigSnapDS, error)
)
