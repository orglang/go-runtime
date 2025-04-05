package proc

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"smecalculus/rolevod/lib/id"

	"smecalculus/rolevod/internal/step"
)

type SpecMsg struct {
	ProcID string       `json:"proc_id" param:"id"`
	PoolID string       `json:"pool_id"`
	Term   step.CallMsg `json:"term"`
}

func (dto SpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.ProcID, id.Required...),
		validation.Field(&dto.PoolID, id.Required...),
		validation.Field(&dto.Term, validation.Required),
	)
}

type IdentMsg struct {
	ProcID string `param:"id"`
}

type RefMsg struct {
	ProcID string `json:"proc_id"`
}

type SnapMsg struct {
	ProcID string `json:"proc_id"`
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/internal/step:Msg.*
var (
	MsgFromSpec func(Spec) SpecMsg
	MsgToSpec   func(SpecMsg) (Spec, error)
	MsgToRef    func(RefMsg) (Ref, error)
	MsgFromRef  func(Ref) RefMsg
	MsgToSnap   func(SnapMsg) (Snap, error)
	MsgFromSnap func(Snap) SnapMsg
)
