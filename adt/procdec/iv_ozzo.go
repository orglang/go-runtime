package procdec

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/orglang/go-sdk/adt/identity"
	"github.com/orglang/go-sdk/adt/qualsym"
)

func (dto DecSpecVP) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.ProcNS, qualsym.Required...),
		validation.Field(&dto.ProcSN, qualsym.Required...),
	)
}

func (dto DecRefVP) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.DecID, identity.Required...),
	)
}
