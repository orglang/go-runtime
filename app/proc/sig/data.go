package sig

type bndSpecData struct {
	ChnlPH string `json:"chnl_ph"`
	RoleQN string `json:"role_qn"`
}

type refData struct {
	SigID string `db:"sig_id"`
	SigRN int64  `db:"rev"`
	Title string `db:"title"`
}

type rootData struct {
	SigID string        `db:"sig_id"`
	Title string        `db:"title"`
	Ys    []bndSpecData `db:"ys"`
	X     bndSpecData   `db:"x"`
	SigRN int64         `db:"rev"`
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/internal/state:Data.*
var (
	DataToRef     func(refData) (Ref, error)
	DataFromRef   func(Ref) refData
	DataToRefs    func([]refData) ([]Ref, error)
	DataFromRefs  func([]Ref) []refData
	DataToRoot    func(rootData) (Impl, error)
	DataFromRoot  func(Impl) (rootData, error)
	DataToRoots   func([]rootData) ([]Impl, error)
	DataFromRoots func([]Impl) ([]rootData, error)
)
