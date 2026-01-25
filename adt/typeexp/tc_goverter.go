package typeexp

import (
	"github.com/orglang/go-sdk/adt/typeexp"
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
// goverter:extend orglang/go-runtime/adt/revnum:Convert.*
// goverter:extend Msg.*
var (
	MsgFromExpRefs func([]ExpRef) []typeexp.ExpRef
	MsgToExpRefs   func([]typeexp.ExpRef) ([]ExpRef, error)
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
// goverter:extend orglang/go-runtime/adt/revnum:Convert.*
// goverter:extend Data.*
var (
	DataToExpRefs   func([]*ExpRefDS) ([]ExpRef, error)
	DataFromExpRefs func([]ExpRef) []*ExpRefDS
	DataToExpRecs   func([]*expRecDS) ([]ExpRec, error)
	DataFromExpRecs func([]ExpRec) []*expRecDS
)
