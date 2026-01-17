package termctx

import (
	"github.com/orglang/go-sdk/adt/termctx"
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/qualsym:Convert.*
var (
	MsgToBindClaim   func(termctx.BindClaimME) (BindClaim, error)
	MsgFromBindClaim func(BindClaim) termctx.BindClaimME
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/qualsym:Convert.*
var (
	DataToBindClaim   func(BindClaimDS) (BindClaim, error)
	DataFromBindClaim func(BindClaim) BindClaimDS
)
