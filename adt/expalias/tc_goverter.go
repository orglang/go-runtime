package expalias

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/adt/identity:Convert.*
// goverter:extend orglang/orglang/adt/revnum:Convert.*
// goverter:extend orglang/orglang/adt/qualsym:Convert.*
var (
	DataFromRoot func(Root) (rootDS, error)
	DataToRoot   func(rootDS) (Root, error)
)
