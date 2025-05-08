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
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/app/type/def:Data.*
var (
	DataToSigRef     func(sigRefData) (SigRef, error)
	DataFromSigRef   func(SigRef) sigRefData
	DataToSigRefs    func([]sigRefData) ([]SigRef, error)
	DataFromSigRefs  func([]SigRef) []sigRefData
	DataToSigRec     func(sigRecData) (SigRec, error)
	DataFromSigRec   func(SigRec) (sigRecData, error)
	DataToSigRecs    func([]sigRecData) ([]SigRec, error)
	DataFromSigRecs  func([]SigRec) ([]sigRecData, error)
	DataToSigSnap    func(sigSnapData) (SigSnap, error)
	DataFromSigSnap  func(SigSnap) (sigSnapData, error)
	DataToSigSnaps   func([]sigSnapData) ([]SigSnap, error)
	DataFromSigSnaps func([]SigSnap) ([]sigSnapData, error)
)
