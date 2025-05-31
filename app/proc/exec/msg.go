package exec

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"smecalculus/rolevod/lib/id"

	procdef "smecalculus/rolevod/app/proc/def"
)

type SpecMsg struct {
	ProcID string              `json:"proc_id" param:"id"`
	PoolID string              `json:"pool_id"`
	Term   procdef.CallSpecMsg `json:"term"`
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
// goverter:extend smecalculus/rolevod/app/proc/def:Msg.*
var (
	MsgFromSpec func(ProcSpec) SpecMsg
	MsgToSpec   func(SpecMsg) (ProcSpec, error)
	MsgToRef    func(RefMsg) (ProcRef, error)
	MsgFromRef  func(ProcRef) RefMsg
	MsgToSnap   func(SnapMsg) (ProcSnap, error)
	MsgFromSnap func(ProcSnap) SnapMsg
)
