package sd

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (dto storagePC) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Protocol, validation.Required),
	)
}

func (dto protocolPC) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Mode, validation.Required, validation.In(postgresMode)),
		validation.Field(&dto.Postgres, validation.Required.When(dto.Mode == postgresMode)),
	)
}

func (dto postgresPC) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Url, validation.Required),
	)
}

func (dto driverPC) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Mode, validation.In(pgxMode)),
		validation.Field(&dto.Pgx, validation.Required.When(dto.Mode == pgxMode)),
	)
}

func (dto pgxPC) Validate() error {
	return validation.ValidateStruct(&dto)
}
