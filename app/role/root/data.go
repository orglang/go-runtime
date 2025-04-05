package role

type refData struct {
	RoleID string `db:"role_id"`
	RoleRN int64  `db:"rev"`
	Title  string `db:"title"`
}

type rootData struct {
	RoleID  string `db:"role_id"`
	Title   string `db:"title"`
	StateID string `db:"state_id"`
	RoleRN  int64  `db:"rev"`
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/lib/rn:Convert.*
// goverter:extend smecalculus/rolevod/internal/state:Data.*
var (
	DataToRef     func(refData) (Ref, error)
	DataFromRef   func(Ref) (refData, error)
	DataToRefs    func([]refData) ([]Ref, error)
	DataFromRefs  func([]Ref) ([]refData, error)
	DataToRoot    func(rootData) (Impl, error)
	DataFromRoot  func(Impl) (rootData, error)
	DataToRoots   func([]rootData) ([]Impl, error)
	DataFromRoots func([]Impl) ([]rootData, error)
)
