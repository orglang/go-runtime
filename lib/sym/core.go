package sym

import (
	"strings"

	"smecalculus/rolevod/lib/ph"
)

var (
	Nil ADT
)

type Symbolizable interface {
	Sym() ADT
}

type ADT string

func New(qn string) ADT {
	return ADT(qn)
}

func (ns ADT) New(sn string) ADT {
	return ADT(strings.Join([]string{string(ns), sn}, sep))
}

// short name or simple name
func (s ADT) SN() string {
	sym := string(s)
	return sym[strings.LastIndex(sym, sep)+1:]
}

// namespace
func (s ADT) NS() ADT {
	sym := string(s)
	return ADT(sym[0:strings.LastIndex(sym, sep)])
}

func (s ADT) ToPH() ph.ADT {
	return ph.ADT(s.SN())
}

func ConvertToSame(a ADT) ADT {
	return a
}

func CovertFromString(s string) ADT {
	return ADT(s)
}

func ConvertToString(s ADT) string {
	return string(s)
}

func ConvertToPH(s ADT) ph.ADT {
	return ph.ADT(s.SN())
}

const (
	sep = "."
)
