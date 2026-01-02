package expctx

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/adt/qualsym:Convert.*
var (
	MsgToBindClaim   func(BindClaimME) (BindClaim, error)
	MsgFromBindClaim func(BindClaim) BindClaimME
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/adt/qualsym:Convert.*
var (
	DataToBindClaim   func(BindClaimDS) (BindClaim, error)
	DataFromBindClaim func(BindClaim) BindClaimDS
)
