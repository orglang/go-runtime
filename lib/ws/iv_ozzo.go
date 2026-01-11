package ws

import (
	"slices"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (dto exchangeCS) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Protocol, validation.Required),
		validation.Field(&dto.Server, validation.Required),
	)
}

func (dto protocolCS) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Modes, validation.Required, validation.Each(validation.In(httpProto))),
		validation.Field(&dto.Http, validation.Required.When(slices.Contains(dto.Modes, httpProto))),
	)
}

func (dto httpCS) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Port, validation.Required, validation.Min(80), validation.Max(20000)),
	)
}

func (dto serverCS) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Mode, validation.Required, validation.In(echoServer)),
		// validation.Field(&dto.Echo, validation.Required.When(dto.Mode == echoMode)),
	)
}

func (dto echoCS) Validate() error {
	return validation.ValidateStruct(&dto)
}
