package typedef

import (
	"github.com/orglang/go-sdk/adt/typedef"
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
// goverter:extend orglang/go-runtime/adt/typedef:Convert.*
var (
	ConvertRecToRef  func(DefRec) DefRef
	ConvertSnapToRef func(DefSnap) DefRef
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
// goverter:extend orglang/go-runtime/adt/revnum:Convert.*
// goverter:extend orglang/go-runtime/adt/typedef:Msg.*
// goverter:extend Msg.*
var (
	MsgFromDefSpec  func(DefSpec) typedef.DefSpecME
	MsgToDefSpec    func(typedef.DefSpecME) (DefSpec, error)
	MsgFromDefRef   func(DefRef) typedef.DefRefME
	MsgToDefRef     func(typedef.DefRefME) (DefRef, error)
	MsgFromDefRefs  func([]DefRef) []typedef.DefRefME
	MsgToDefRefs    func([]typedef.DefRefME) ([]DefRef, error)
	MsgFromDefSnap  func(DefSnap) typedef.DefSnapME
	MsgToDefSnap    func(typedef.DefSnapME) (DefSnap, error)
	MsgFromDefSnaps func([]DefSnap) []typedef.DefSnapME
	MsgToDefSnaps   func([]typedef.DefSnapME) ([]DefSnap, error)
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
// goverter:extend orglang/go-runtime/adt/revnum:Convert.*
// goverter:extend orglang/go-runtime/adt/typedef:Msg.*
var (
	ViewFromDefRef  func(DefRef) DefRefVP
	ViewToDefRef    func(DefRefVP) (DefRef, error)
	ViewFromDefRefs func([]DefRef) []DefRefVP
	ViewToDefRefs   func([]DefRefVP) ([]DefRef, error)
	ViewFromDefSnap func(DefSnap) DefSnapVP
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
// goverter:extend orglang/go-runtime/adt/revnum:Convert.*
// goverter:extend orglang/go-runtime/adt/typedef:Data.*
// goverter:extend data.*
// goverter:extend DataToTermRef
// goverter:extend DataFromTermRef
var (
	DataToDefRef    func(defRefDS) (DefRef, error)
	DataFromDefRef  func(DefRef) (defRefDS, error)
	DataToDefRefs   func([]defRefDS) ([]DefRef, error)
	DataFromDefRefs func([]DefRef) ([]defRefDS, error)
	DataToDefRec    func(defRecDS) (DefRec, error)
	DataFromDefRec  func(DefRec) (defRecDS, error)
	DataToDefRecs   func([]defRecDS) ([]DefRec, error)
	DataFromDefRecs func([]DefRec) ([]defRecDS, error)
)
