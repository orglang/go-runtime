package dec

type bndSpecData struct {
	ChnlPH string `json:"chnl_ph"`
	TypeQN string `json:"role_qn"`
}

type sigRefData struct {
	SigID string `db:"sig_id"`
	SigRN int64  `db:"rev"`
	Title string `db:"title"`
}

type sigRecData struct {
	SigID string        `db:"sig_id"`
	Title string        `db:"title"`
	Ys    []bndSpecData `db:"ys"`
	X     bndSpecData   `db:"x"`
	SigRN int64         `db:"rev"`
}

type sigSnapData struct {
	SigID string        `db:"sig_id"`
	Title string        `db:"title"`
	Ys    []bndSpecData `db:"ys"`
	X     bndSpecData   `db:"x"`
	SigRN int64         `db:"rev"`
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/avt/id:Convert.*
// goverter:extend orglang/orglang/aat/type/def:Data.*
var (
	DataToSigRef     func(sigRefData) (ProcRef, error)
	DataFromSigRef   func(ProcRef) sigRefData
	DataToSigRefs    func([]sigRefData) ([]ProcRef, error)
	DataFromSigRefs  func([]ProcRef) []sigRefData
	DataToSigRec     func(sigRecData) (ProcRec, error)
	DataFromSigRec   func(ProcRec) (sigRecData, error)
	DataToSigRecs    func([]sigRecData) ([]ProcRec, error)
	DataFromSigRecs  func([]ProcRec) ([]sigRecData, error)
	DataToSigSnap    func(sigSnapData) (ProcSnap, error)
	DataFromSigSnap  func(ProcSnap) (sigSnapData, error)
	DataToSigSnaps   func([]sigSnapData) ([]ProcSnap, error)
	DataFromSigSnaps func([]ProcSnap) ([]sigSnapData, error)
)
