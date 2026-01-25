package revnum

type ADT int64

func New() ADT {
	return ADT(1)
}

func Next(rn ADT) ADT {
	return rn + 1
}

func (rn ADT) Next() ADT {
	return rn + 1
}
