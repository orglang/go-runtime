package procdec

import "github.com/orglang/go-sdk/adt/uniqref"

type DecRefVP = uniqref.Msg

type DecSpecVP struct {
	ProcQN string `form:"qn" json:"qn"`
}

type DecSnapVP struct {
	DecRef DecRefVP `json:"ref"`
}
