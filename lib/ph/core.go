package ph

var (
	Nil ADT
)

type ADT string

func New(s string) ADT {
	return ADT(s)
}
