package ph

func DataFromPH(ph ADT) string {
	return string(ph)
}

func DataToPH(dto string) (ADT, error) {
	return ADT(dto), nil
}
