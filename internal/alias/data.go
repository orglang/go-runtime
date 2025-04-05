package alias

type rootData struct {
	ID  string
	RN  int64
	Sym string
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/lib/rn:Convert.*
// goverter:extend smecalculus/rolevod/lib/sym:Convert.*
var (
	DataFromRoot func(Root) (rootData, error)
	DataToRoot   func(rootData) (Root, error)
)
