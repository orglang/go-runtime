package dec

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"orglang/orglang/avt/core"
	"orglang/orglang/avt/id"
	"orglang/orglang/avt/sym"
)

func (dto BndSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.ChnlPH, sym.Optional...),
		validation.Field(&dto.TypeQN, sym.Required...),
	)
}

func (dto SigSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.SigQN, sym.Required...),
		validation.Field(&dto.Ys, core.CtxOptional...),
	)
}

func (dto IdentME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.SigID, id.Required...),
	)
}

func (dto SigRefME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.SigID, id.Required...),
	)
}

func (dto SigSpecVP) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.SigNS, sym.Required...),
		validation.Field(&dto.SigSN, sym.Required...),
	)
}

func (dto SigRefVP) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.SigID, id.Required...),
		validation.Field(&dto.Title, sym.Required...),
	)
}
