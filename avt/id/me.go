package id

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var Optional = []validation.Rule{
	validation.Length(20, 20),
	validation.Match(regexp.MustCompile(`^[0-9a-z]*$`)),
}

var Required = append(Optional, validation.Required)

func RequiredWhen(condition bool) []validation.Rule {
	return append(Optional, validation.Required.When(condition))
}
