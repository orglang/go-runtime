package typedef

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/orglang/go-sdk/adt/identity"
	"github.com/orglang/go-sdk/adt/revnum"
	"github.com/orglang/go-sdk/adt/uniqsym"
)

func (dto DefSpecVP) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.TypeNS, uniqsym.Required...),
		validation.Field(&dto.TypeSN, uniqsym.Required...),
	)
}

func (dto DefRefVP) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.DefID, identity.Required...),
		validation.Field(&dto.DefRN, revnum.Optional...),
		validation.Field(&dto.Title, uniqsym.Required...),
	)
}
