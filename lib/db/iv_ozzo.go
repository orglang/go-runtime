package db

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (dto storageCS) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Protocol, validation.Required),
		validation.Field(&dto.Driver, validation.Required),
	)
}

func (dto protocolCS) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Mode, validation.Required, validation.In(postgresProto)),
		validation.Field(&dto.Postgres, validation.Required.When(dto.Mode == postgresProto)),
	)
}

func (dto postgresCS) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Url, validation.Required),
	)
}

func (dto driverCS) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Mode, validation.Required, validation.In(pgxDriver)),
		// validation.Field(&dto.Pgx, validation.Required.When(dto.Mode == pgxMode)),
	)
}

func (dto pgxCS) Validate() error {
	return validation.ValidateStruct(&dto)
}
