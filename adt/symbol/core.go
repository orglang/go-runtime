package symbol

type ADT string

func New(str string) ADT {
	if str == "" {
		panic("invalid symbol")
	}
	return ADT(str)
}
