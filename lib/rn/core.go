package rn

type ADT int64

func Initial() ADT {
	return ADT(1)
}

func Next(rev ADT) ADT {
	return rev + 1
}

func (rev ADT) Next() ADT {
	return rev + 1
}

func ConvertToInt(a ADT) int64 {
	return int64(a)
}

func ConvertFromInt(i int64) ADT {
	return ADT(i)
}
