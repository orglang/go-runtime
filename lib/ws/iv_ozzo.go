package ws

import (
	"slices"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (dto exchangePC) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Protocol, validation.Required),
		validation.Field(&dto.Server, validation.Required),
	)
}

func (dto protocolPC) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Modes, validation.Required, validation.Each(validation.In(httpMode))),
		validation.Field(&dto.Http, validation.Required.When(slices.Contains(dto.Modes, httpMode))),
	)
}

func (dto httpPC) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Port, validation.Required, validation.Min(80), validation.Max(20000)),
	)
}

func (dto serverPC) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Mode, validation.Required, validation.In(echoMode)),
		// validation.Field(&dto.Echo, validation.Required.When(dto.Mode == echoMode)),
	)
}

func (dto echoPC) Validate() error {
	return validation.ValidateStruct(&dto)
}
