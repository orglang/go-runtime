package procexp

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/qualsym"
	"orglang/orglang/adt/termctx"
)

var expKindRequired = []validation.Rule{
	validation.Required,
	validation.In(Close, Wait, Send, Recv, Lab, Case, Spawn, Fwd, Call),
}

func (dto ExpSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.K, expKindRequired...),
		validation.Field(&dto.Close, validation.Required.When(dto.K == Close)),
		validation.Field(&dto.Wait, validation.Required.When(dto.K == Wait)),
		validation.Field(&dto.Send, validation.Required.When(dto.K == Send)),
		validation.Field(&dto.Recv, validation.Required.When(dto.K == Recv)),
		validation.Field(&dto.Lab, validation.Required.When(dto.K == Lab)),
		validation.Field(&dto.Case, validation.Required.When(dto.K == Case)),
		validation.Field(&dto.Spawn, validation.Required.When(dto.K == Spawn)),
		validation.Field(&dto.Fwd, validation.Required.When(dto.K == Fwd)),
		validation.Field(&dto.Call, validation.Required.When(dto.K == Call)),
	)
}

func (dto CloseSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.CommPH, validation.Required),
	)
}

func (dto WaitSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.CommPH, validation.Required),
		validation.Field(&dto.ContES, validation.Required),
	)
}

func (dto SendSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.CommPH, validation.Required),
		validation.Field(&dto.ValPH, validation.Required),
	)
}

func (dto RecvSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.CommPH, validation.Required),
		validation.Field(&dto.BindPH, validation.Required),
		validation.Field(&dto.ContES, validation.Required),
	)
}

func (dto LabSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.CommPH, validation.Required),
		validation.Field(&dto.Label, qualsym.Required...),
	)
}

func (dto CaseSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.CommPH, validation.Required),
		validation.Field(&dto.ContBSs,
			validation.Required,
			validation.Length(1, 10),
			validation.Each(validation.Required),
		),
	)
}

func (dto BranchSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Label, qualsym.Required...),
		validation.Field(&dto.ContES, validation.Required),
	)
}

func (dto CallSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.CommPH, validation.Required),
		validation.Field(&dto.ProcQN, identity.Required...),
		validation.Field(&dto.ValPHs, termctx.Optional...),
	)
}

func (dto SpawnSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.CommPH, validation.Required),
		validation.Field(&dto.DecID, identity.Required...),
		validation.Field(&dto.BindPHs, termctx.Optional...),
	)
}

func (dto FwdSpecME) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.Y, validation.Required),
	)
}
