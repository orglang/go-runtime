package rn

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var Optional = []validation.Rule{
	validation.Min(0),
}

var Required = append(Optional, validation.Required)
