package typedef

import (
	"github.com/orglang/go-sdk/adt/typeexp"
	"github.com/orglang/go-sdk/adt/uniqref"
)

type DefSpecVP struct {
	TypeNS string `form:"ns" json:"ns"`
	TypeSN string `form:"name" json:"name"`
}

type DefRefVP = uniqref.Msg

type DefSnapVP struct {
	DefRef DefRefVP        `json:"ref"`
	Title  string          `json:"title"`
	TypeES typeexp.ExpSpec `json:"type_es"`
}
