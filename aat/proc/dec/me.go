package dec

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"orglang/orglang/avt/core"
	"orglang/orglang/avt/id"
	"orglang/orglang/avt/sym"
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
// goverter:extend orglang/orglang/avt/sym:Convert.*
var (
	MsgToBndSpec   func(BndSpecMsg) (ChnlSpec, error)
	MsgFromBndSpec func(ChnlSpec) BndSpecMsg
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
// goverter:extend orglang/orglang/avt/id:Convert.*
// goverter:extend orglang/orglang/aat/type/def:Msg.*
var (
	MsgToSigSpec    func(SigSpecMsg) (ProcSpec, error)
	MsgFromSigSpec  func(ProcSpec) SigSpecMsg
	MsgToSigRef     func(SigRefMsg) (ProcRef, error)
	MsgFromSigRef   func(ProcRef) SigRefMsg
	MsgToSigSnap    func(SigSnapMsg) (ProcSnap, error)
	MsgFromSigSnap  func(ProcSnap) SigSnapMsg
	MsgFromSigSnaps func([]ProcSnap) []SigSnapMsg
)
