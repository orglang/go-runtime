package procdec

import (
	"github.com/orglang/go-sdk/adt/procdec"
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/uniqsym:Convert.*
// goverter:extend orglang/go-runtime/adt/procbind:Msg.*
// goverter:extend orglang/go-runtime/adt/typedef:Msg.*
var (
	MsgToDecSpec    func(procdec.DecSpec) (DecSpec, error)
	MsgFromDecSpec  func(DecSpec) procdec.DecSpec
	MsgToDecSnap    func(procdec.DecSnap) (DecSnap, error)
	MsgFromDecSnap  func(DecSnap) procdec.DecSnap
	MsgFromDecSnaps func([]DecSnap) []procdec.DecSnap
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/uniqsym:Convert.*
// goverter:extend orglang/go-runtime/adt/uniqref:Msg.*
var (
	ViewFromDecSnap func(DecSnap) DecSnapVP
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
// goverter:extend orglang/go-runtime/adt/procbind:Data.*
var (
	// goverter:map . DecRef
	DataToDecRec func(decRecDS) (DecRec, error)
	// goverter:autoMap DecRef
	DataFromDecRec  func(DecRec) (decRecDS, error)
	DataToDecRecs   func([]decRecDS) ([]DecRec, error)
	DataFromDecRecs func([]DecRec) ([]decRecDS, error)
	// goverter:map . DecRef
	DataToDecSnap func(decSnapDS) (DecSnap, error)
	// goverter:autoMap DecRef
	DataFromDecSnap  func(DecSnap) (decSnapDS, error)
	DataToDecSnaps   func([]decSnapDS) ([]DecSnap, error)
	DataFromDecSnaps func([]DecSnap) ([]decSnapDS, error)
)
