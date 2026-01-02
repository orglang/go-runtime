package procdecl

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"orglang/orglang/adt/expctx"
	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/qualsym"
)

func (dto SigSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.SigQN, qualsym.Required...),
		validation.Field(&dto.Ys, expctx.Optional...),
	)
}

func (dto IdentME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.SigID, identity.Required...),
	)
}

func (dto SigRefME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.SigID, identity.Required...),
	)
}

func (dto SigSpecVP) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.SigNS, qualsym.Required...),
		validation.Field(&dto.SigSN, qualsym.Required...),
	)
}

func (dto SigRefVP) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.SigID, identity.Required...),
		validation.Field(&dto.Title, qualsym.Required...),
	)
}
