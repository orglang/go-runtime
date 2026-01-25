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
// goverter:extend orglang/go-runtime/adt/uniqref:Msg.*
// goverter:extend orglang/go-runtime/adt/typeexp:Msg.*
// goverter:extend Msg.*
var (
	MsgFromDefSpec  func(DefSpec) typedef.DefSpec
	MsgToDefSpec    func(typedef.DefSpec) (DefSpec, error)
	MsgFromDefRef   func(DefRef) typedef.DefRef
	MsgToDefRef     func(typedef.DefRef) (DefRef, error)
	MsgFromDefRefs  func([]DefRef) []typedef.DefRef
	MsgToDefRefs    func([]typedef.DefRef) ([]DefRef, error)
	MsgFromDefSnap  func(DefSnap) typedef.DefSnap
	MsgToDefSnap    func(typedef.DefSnap) (DefSnap, error)
	MsgFromDefSnaps func([]DefSnap) []typedef.DefSnap
	MsgToDefSnaps   func([]typedef.DefSnap) ([]DefSnap, error)
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/uniqref:Msg.*
// goverter:extend orglang/go-runtime/adt/typeexp:Msg.*
var (
	ViewFromDefRef  func(DefRef) DefRefVP
	ViewToDefRef    func(DefRefVP) (DefRef, error)
	ViewFromDefRefs func([]DefRef) []DefRefVP
	ViewToDefRefs   func([]DefRefVP) ([]DefRef, error)
	ViewFromDefSnap func(DefSnap) DefSnapVP
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/uniqref:Data.*
// goverter:extend orglang/go-runtime/adt/typeexp:Data.*
// goverter:extend Data.*
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
