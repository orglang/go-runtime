package dec

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"smecalculus/rolevod/lib/core"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"
)

type BndSpecMsg struct {
	ChnlPH string `json:"chnl_ph"`
	TypeQN string `json:"type_qn"`
}

func (dto BndSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.ChnlPH, sym.Optional...),
		validation.Field(&dto.TypeQN, sym.Required...),
	)
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/sym:Convert.*
var (
	MsgToBndSpec   func(BndSpecMsg) (BndSpec, error)
	MsgFromBndSpec func(BndSpec) BndSpecMsg
)

type SigSpecMsg struct {
	X     BndSpecMsg   `json:"x"`
	SigQN string       `json:"sig_qn"`
	Ys    []BndSpecMsg `json:"ys"`
}

func (dto SigSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.SigQN, sym.Required...),
		validation.Field(&dto.Ys, core.CtxOptional...),
	)
}

type IdentMsg struct {
	SigID string `json:"id" param:"id"`
}

func (dto IdentMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.SigID, id.Required...),
	)
}

type SigRefMsg struct {
	SigID string `json:"id" param:"id"`
	Title string `json:"title"`
	SigRN int64  `json:"rev"`
}

func (dto SigRefMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.SigID, id.Required...),
	)
}

type SigSnapMsg struct {
	X     BndSpecMsg   `json:"x"`
	SigID string       `json:"sig_id"`
	Ys    []BndSpecMsg `json:"ys"`
	Title string       `json:"title"`
	SigRN int64        `json:"sig_rn"`
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/app/type/def:Msg.*
var (
	MsgToSigSpec    func(SigSpecMsg) (SigSpec, error)
	MsgFromSigSpec  func(SigSpec) SigSpecMsg
	MsgToSigRef     func(SigRefMsg) (SigRef, error)
	MsgFromSigRef   func(SigRef) SigRefMsg
	MsgToSigSnap    func(SigSnapMsg) (SigSnap, error)
	MsgFromSigSnap  func(SigSnap) SigSnapMsg
	MsgFromSigSnaps func([]SigSnap) []SigSnapMsg
)
