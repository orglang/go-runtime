package syndec

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
// goverter:extend orglang/go-runtime/adt/uniqsym:Convert.*
// goverter:extend orglang/go-runtime/adt/revnum:Convert.*
var (
	DataFromDecRec func(DecRec) (decRecDS, error)
	DataToDecRec   func(decRecDS) (DecRec, error)
)
