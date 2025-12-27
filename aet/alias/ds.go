package alias

type rootData struct {
	ID  string
	RN  int64
	Sym string
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/avt/id:Convert.*
// goverter:extend orglang/orglang/avt/rn:Convert.*
// goverter:extend orglang/orglang/avt/sym:Convert.*
var (
	DataFromRoot func(Root) (rootData, error)
	DataToRoot   func(rootData) (Root, error)
)
