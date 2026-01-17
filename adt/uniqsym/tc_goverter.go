package uniqsym

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend Convert.*
var (
	ConvertFromStrings func([]string) ([]ADT, error)
	ConvertToStrings   func([]ADT) []string
)
