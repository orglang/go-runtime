package dec

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"orglang/orglang/avt/id"
	"orglang/orglang/avt/sym"
)

type SigSpecView struct {
	SigNS string `form:"ns" json:"ns"`
	SigSN string `form:"sn" json:"sn"`
}

func (dto SigSpecView) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.SigNS, sym.Required...),
		validation.Field(&dto.SigSN, sym.Required...),
	)
}

type SigRefView struct {
	SigID string `form:"sig_id" json:"sig_id" param:"id"`
	SigRN int64  `form:"sig_rn" json:"sig_rn"`
	Title string `form:"name" json:"title"`
}

func (dto SigRefView) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.SigID, id.Required...),
		validation.Field(&dto.Title, sym.Required...),
	)
}

type SigSnapView struct {
	SigID string `json:"sig_id"`
	SigRN int64  `json:"sig_rn"`
	Title string `json:"title"`
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/avt/id:Convert.*
// goverter:extend orglang/orglang/avt/rn:Convert.*
// goverter:extend orglang/orglang/aat/type/def:Msg.*
// goverter:extend Msg.*
var (
	ViewFromSigRef  func(ProcRef) SigRefView
	ViewToSigRef    func(SigRefView) (ProcRef, error)
	ViewFromSigRefs func([]ProcRef) []SigRefView
	ViewToSigRefs   func([]SigRefView) ([]ProcRef, error)
	ViewFromSigSnap func(ProcSnap) SigSnapView
)
