package symbol

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend ConvertToString
// goverter:extend ConvertFromString
var (
	ConvertFromStrings func([]string) ([]ADT, error)
	ConvertToStrings   func([]ADT) []string
)
