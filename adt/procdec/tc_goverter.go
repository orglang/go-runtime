package procdec

import "github.com/orglang/go-sdk/adt/procdec"

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
var (
	ConvertSnapToRef func(DecSnap) DecRef
	ConvertRecToRef  func(DecRec) DecRef
	ConvertRecToSnap func(DecRec) DecSnap
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/qualsym:Convert.*
// goverter:extend orglang/go-runtime/adt/termctx:Msg.*
// goverter:extend orglang/go-runtime/adt/typedef:Msg.*
var (
	MsgToDecSpec    func(procdec.DecSpecME) (DecSpec, error)
	MsgFromDecSpec  func(DecSpec) procdec.DecSpecME
	MsgToDecRef     func(procdec.DecRefME) (DecRef, error)
	MsgFromDecRef   func(DecRef) procdec.DecRefME
	MsgToDecSnap    func(procdec.DecSnapME) (DecSnap, error)
	MsgFromDecSnap  func(DecSnap) procdec.DecSnapME
	MsgFromDecSnaps func([]DecSnap) []procdec.DecSnapME
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
// goverter:extend orglang/go-runtime/adt/revnum:Convert.*
// goverter:extend orglang/go-runtime/adt/typedef:Msg.*
// goverter:extend Msg.*
var (
	ViewFromDecRef  func(DecRef) DecRefVP
	ViewToDecRef    func(DecRefVP) (DecRef, error)
	ViewFromDecRefs func([]DecRef) []DecRefVP
	ViewToDecRefs   func([]DecRefVP) ([]DecRef, error)
	ViewFromDecSnap func(DecSnap) DecSnapVP
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
// goverter:extend orglang/go-runtime/adt/termctx:Data.*
// goverter:extend orglang/go-runtime/adt/typedef:Data.*
var (
	DataToDecRef     func(decRefDS) (DecRef, error)
	DataFromDecRef   func(DecRef) decRefDS
	DataToDecRefs    func([]decRefDS) ([]DecRef, error)
	DataFromDecRefs  func([]DecRef) []decRefDS
	DataToDecRec     func(decRecDS) (DecRec, error)
	DataFromDecRec   func(DecRec) (decRecDS, error)
	DataToDecRecs    func([]decRecDS) ([]DecRec, error)
	DataFromDecRecs  func([]DecRec) ([]decRecDS, error)
	DataToDecSnap    func(decSnapDS) (DecSnap, error)
	DataFromDecSnap  func(DecSnap) (decSnapDS, error)
	DataToDecSnaps   func([]decSnapDS) ([]DecSnap, error)
	DataFromDecSnaps func([]DecSnap) ([]decSnapDS, error)
)
