package qualsym

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

func (s ADT) String() string {
	return string(s)
}

const (
	sep = "."
)
