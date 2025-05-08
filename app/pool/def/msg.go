package def

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"smecalculus/rolevod/lib/id"

	procdef "smecalculus/rolevod/app/proc/def"
)

type PoolSpecMsg struct {
	SigQN   string   `json:"sig_qn"`
	ProcIDs []string `json:"proc_ids"`
	SupID   string   `json:"suo_id"`
}

func (dto PoolSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.SupID, id.Optional...),
	)
}

type IdentMsg struct {
	PoolID string `json:"id" param:"id"`
}

type PoolRefMsg struct {
	PoolID string `json:"pool_id"`
	ProcID string `json:"proc_id"`
}

type PoolSnapMsg struct {
	PoolID string       `json:"id"`
	Title  string       `json:"title"`
	Subs   []PoolRefMsg `json:"subs"`
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
var (
	MsgToPoolSpec   func(PoolSpecMsg) (PoolSpec, error)
	MsgFromPoolSpec func(PoolSpec) PoolSpecMsg
	MsgToPoolRef    func(PoolRefMsg) (PoolRef, error)
	MsgFromPoolRef  func(PoolRef) PoolRefMsg
	MsgToPoolSnap   func(PoolSnapMsg) (PoolSnap, error)
	MsgFromPoolSnap func(PoolSnap) PoolSnapMsg
)

type StepSpecMsg struct {
	PoolID string              `json:"pool_id"`
	ProcID string              `json:"proc_id"`
	Term   procdef.TermSpecMsg `json:"term"`
}

func (dto StepSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.PoolID, id.Required...),
		validation.Field(&dto.ProcID, id.Required...),
		validation.Field(&dto.Term, validation.Required),
	)
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/app/proc/def:Msg.*
var (
	MsgFromStepSpec func(StepSpec) StepSpecMsg
	MsgToStepSpec   func(StepSpecMsg) (StepSpec, error)
)
