package uniqsym

import (
	"fmt"
	"strings"

	"orglang/go-runtime/adt/symbol"
)

const (
	sep = "."
)

func ConvertFromString(str string) (ADT, error) {
	if str == "" {
		return empty, fmt.Errorf("invalid value")
	}
	idx := strings.LastIndex(str, sep)
	if idx < 0 {
		return ADT{symbol.New(str), nil}, nil
	}
	name, err := symbol.ConvertFromString(str[idx+1:])
	if err != nil {
		return empty, err
	}
	space, err := ConvertFromString(str[:idx])
	if err != nil {
		return empty, err
	}
	return ADT{name, &space}, nil
}

func ConvertToString(adt ADT) string {
	if adt == empty {
		panic("invalid value")
	}
	name := symbol.ConvertToString(adt.sym)
	if adt.ns == nil {
		return name
	}
	space := ConvertToString(*adt.ns)
	return space + sep + name
}
