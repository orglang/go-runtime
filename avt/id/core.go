package id

import (
	"errors"

	"github.com/rs/xid"
)

var (
	Nil ADT
)

type Identifiable interface {
	Ident() ADT
}

type ADT xid.ID

func (ADT) PH() {}

func New() ADT {
	return ADT(xid.New())
}

func Empty() ADT {
	return ADT(xid.NilID())
}

func (id ADT) IsEmpty() bool {
	return xid.ID(id).IsZero()
}

func (id ADT) String() string {
	return xid.ID(id).String()
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend ConvertFromString
// goverter:extend ConvertToString
var (
	ConvertFromStrings func([]string) ([]ADT, error)
	ConvertToStrings   func([]ADT) []string
)

func ConvertToSame(id ADT) ADT {
	return id
}

func ConvertFromString(s string) (ADT, error) {
	xid, err := xid.FromString(s)
	if err != nil {
		return ADT{}, err
	}
	return ADT(xid), nil
}

func ConvertToString(id ADT) string {
	return xid.ID(id).String()
}

func ConvertPtrToStringPtr(id *ADT) *string {
	if id == nil {
		return nil
	}
	s := xid.ID(*id).String()
	return &s
}

var (
	ErrEmpty = errors.New("empty id")
)
