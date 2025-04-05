package pool

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"smecalculus/rolevod/lib/id"

	"smecalculus/rolevod/internal/step"
)

type SpecMsg struct {
	SigQN   string   `json:"sig_qn"`
	ProcIDs []string `json:"proc_ids"`
	SupID   string   `json:"suo_id"`
}

func (dto SpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.SupID, id.Optional...),
	)
}

type IdentMsg struct {
	PoolID string `json:"id" param:"id"`
}

type RefMsg struct {
	PoolID string `json:"pool_id"`
	ProcID string `json:"proc_id"`
}

type SnapMsg struct {
	PoolID string   `json:"id"`
	Title  string   `json:"title"`
	Subs   []RefMsg `json:"subs"`
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
var (
	MsgToSpec   func(SpecMsg) (Spec, error)
	MsgFromSpec func(Spec) SpecMsg
	MsgToRef    func(RefMsg) (Ref, error)
	MsgFromRef  func(Ref) RefMsg
	MsgToSnap   func(SnapMsg) (Snap, error)
	MsgFromSnap func(Snap) SnapMsg
)

type StepSpecMsg struct {
	PoolID string       `json:"pool_id"`
	ProcID string       `json:"proc_id"`
	Term   step.TermMsg `json:"term"`
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
// goverter:extend smecalculus/rolevod/internal/step:Msg.*
var (
	MsgFromStepSpec func(StepSpec) StepSpecMsg
	MsgToStepSpec   func(StepSpecMsg) (StepSpec, error)
)
