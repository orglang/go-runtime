package sym

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var Optional = []validation.Rule{
	validation.Length(1, 512),
	validation.Match(regexp.MustCompile(`^[0-9A-Za-z_-]*$`)),
}

var Required = append(Optional, validation.Required)

func ReqiredWhen(condition bool) []validation.Rule {
	return append(Optional, validation.Required.When(condition))
}
