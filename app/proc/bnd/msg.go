package bnd

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"smecalculus/rolevod/lib/sym"
)

type SpecMsg struct {
	ChnlPH string `json:"chnl_ph"`
	RoleQN string `json:"role_qn"`
}

func (dto SpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.ChnlPH, sym.Optional...),
		validation.Field(&dto.RoleQN, sym.Required...),
	)
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/ph:Convert.*
// goverter:extend smecalculus/rolevod/lib/sym:Convert.*
var (
	MsgToBndSpec   func(SpecMsg) (Spec, error)
	MsgFromBndSpec func(Spec) SpecMsg
)
