package typedef

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/orglang/go-sdk/adt/identity"
	"github.com/orglang/go-sdk/adt/qualsym"
	"github.com/orglang/go-sdk/adt/revnum"
)

func (dto DefSpecVP) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.TypeNS, qualsym.Required...),
		validation.Field(&dto.TypeSN, qualsym.Required...),
	)
}

func (dto DefRefVP) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.DefID, identity.Required...),
		validation.Field(&dto.DefRN, revnum.Optional...),
		validation.Field(&dto.Title, qualsym.Required...),
	)
}
