package ph

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func ConvertToString(ph ADT) string {
	return string(ph)
}

func ConvertFromString(dto string) (ADT, error) {
	return ADT(dto), nil
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend ConvertToString
// goverter:extend ConvertFromString
var (
	ConvertFromStrings func([]string) ([]ADT, error)
	ConvertToStrings   func([]ADT) []string
)

var Optional = []validation.Rule{
	validation.Length(1, 64),
	validation.Match(regexp.MustCompile(`^[0-9a-zA-Z_-]*$`)),
}

var Required = append(Optional, validation.Required)
