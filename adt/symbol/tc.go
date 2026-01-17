package symbol

import "fmt"

func ConvertFromString(str string) (ADT, error) {
	if str == "" {
		return ADT(""), fmt.Errorf("invalid value")
	}
	return ADT(str), nil
}

func ConvertToString(adt ADT) string {
	return string(adt)
}
