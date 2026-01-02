package expctx

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"orglang/orglang/adt/qualsym"
)

var Optional = []validation.Rule{
	validation.Length(1, 10),
	validation.Each(validation.Required),
}

func (dto BindClaimME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.BindPH, qualsym.Optional...),
		validation.Field(&dto.TypeQN, qualsym.Required...),
	)
}
