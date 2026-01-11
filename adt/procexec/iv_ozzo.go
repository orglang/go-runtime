package procexec

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"orglang/orglang/adt/identity"
)

func (dto ExecSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.ExecID, identity.Required...),
		validation.Field(&dto.PoolID, identity.Required...),
		validation.Field(&dto.ProcES, validation.Required),
	)
}
