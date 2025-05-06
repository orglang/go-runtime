package sym

import (
	"strings"
)

var (
	Nil   ADT
	Blank ADT
)

type Symbolizable interface {
	Sym() ADT
}

type ADT string

func New(s string) ADT {
	return ADT(s)
}

func (ns ADT) New(sn string) ADT {
	return ADT(strings.Join([]string{string(ns), sn}, sep))
}

// simple name
func (s ADT) SN() string {
	sym := string(s)
	return sym[strings.LastIndex(sym, sep)+1:]
}

// namespace
func (s ADT) NS() ADT {
	sym := string(s)
	return ADT(sym[0:strings.LastIndex(sym, sep)])
}

func ConvertToSame(s ADT) ADT {
	return s
}

func ConvertFromString(s string) (ADT, error) {
	return ADT(s), nil
}

func ConvertToString(s ADT) string {
	return string(s)
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend ConvertToString
// goverter:extend ConvertFromString
var (
	ConvertFromStrings func([]string) ([]ADT, error)
	ConvertToStrings   func([]ADT) []string
)

const (
	sep = "."
)
