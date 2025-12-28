package def

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"orglang/orglang/avt/core"
	"orglang/orglang/avt/id"
	"orglang/orglang/avt/rn"
	"orglang/orglang/avt/sym"
)

func (dto TypeSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.TypeQN, sym.Required...),
		validation.Field(&dto.TypeTS, validation.Required),
	)
}

func (dto IdentME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.ID, id.Required...),
	)
}

func (dto TypeRefME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.TypeID, id.Required...),
		validation.Field(&dto.TypeRN, rn.Optional...),
	)
}

func (dto TypeSnapME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.TypeID, id.Required...),
		validation.Field(&dto.TypeRN, rn.Optional...),
		validation.Field(&dto.TypeTS, validation.Required),
	)
}

func (dto TermSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.K, kindRequired...),
		validation.Field(&dto.Link, validation.Required.When(dto.K == LinkKind), validation.Skip.When(dto.K != LinkKind)),
		validation.Field(&dto.Tensor, validation.Required.When(dto.K == TensorKind), validation.Skip.When(dto.K != TensorKind)),
		validation.Field(&dto.Lolli, validation.Required.When(dto.K == LolliKind), validation.Skip.When(dto.K != LolliKind)),
		validation.Field(&dto.Plus, validation.Required.When(dto.K == PlusKind), validation.Skip.When(dto.K != PlusKind)),
		validation.Field(&dto.With, validation.Required.When(dto.K == WithKind), validation.Skip.When(dto.K != WithKind)),
	)
}

func (dto LinkSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.QN, sym.Required...),
	)
}

func (dto ProdSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Value, validation.Required),
		validation.Field(&dto.Cont, validation.Required),
	)
}

func (dto SumSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Choices,
			validation.Required,
			validation.Length(1, 10),
			validation.Each(validation.Required),
		),
	)
}

func (dto ChoiceSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Label, core.NameRequired...),
		validation.Field(&dto.Cont, validation.Required),
	)
}

func (dto TermRefME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.ID, id.Required...),
		validation.Field(&dto.K, kindRequired...),
	)
}

var kindRequired = []validation.Rule{
	validation.Required,
	validation.In(OneKind, LinkKind, TensorKind, LolliKind, PlusKind, WithKind),
}

func (dto TypeSpecVP) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.NS, sym.Required...),
		validation.Field(&dto.Name, sym.Required...),
	)
}

func (dto TypeRefVP) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.RoleID, id.Required...),
		validation.Field(&dto.RoleRN, rn.Optional...),
		validation.Field(&dto.Title, sym.Required...),
	)
}
