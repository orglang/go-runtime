package ph

var (
	Nil   ADT
	Blank ADT
)

type ADT string

func New(s string) ADT {
	return ADT(s)
}
