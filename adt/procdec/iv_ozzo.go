package procdec

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/orglang/go-sdk/adt/uniqsym"
)

func (dto DecSpecVP) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.ProcQN, uniqsym.Required...),
	)
}
