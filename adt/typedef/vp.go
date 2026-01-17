package typedef

import (
	"github.com/orglang/go-sdk/adt/typeexp"
)

type DefSpecVP struct {
	TypeNS string `form:"ns" json:"ns"`
	TypeSN string `form:"name" json:"name"`
}

type DefRefVP struct {
	DefID string `form:"id" json:"id" param:"id"`
	DefRN int64  `form:"def_rn" json:"def_rn"`
	Title string `form:"name" json:"title"`
}

type DefSnapVP struct {
	DefID  string            `json:"def_id"`
	DefRN  int64             `json:"def_rn"`
	Title  string            `json:"title"`
	TypeES typeexp.ExpSpecME `json:"type_es"`
}
