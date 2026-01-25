package procbind

import (
	"github.com/orglang/go-sdk/adt/procbind"
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/symbol:Convert.*
// goverter:extend orglang/go-runtime/adt/uniqsym:Convert.*
var (
	MsgToBindSpec   func(procbind.BindSpec) (BindSpec, error)
	MsgFromBindSpec func(BindSpec) procbind.BindSpec
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
// goverter:extend orglang/go-runtime/adt/uniqsym:Convert.*
var (
	DataToBindSpec   func(BindSpecDS) (BindSpec, error)
	DataFromBindSpec func(BindSpec) BindSpecDS
	// goverter:map . ExecRef
	DataToBindRec func(BindRecDS) (BindRec, error)
	// goverter:autoMap ExecRef
	DataFromBindRec  func(BindRec) BindRecDS
	DataToBindRecs   func([]BindRecDS) ([]BindRec, error)
	DataFromBindRecs func([]BindRec) []BindRecDS
)
