package procdecl

import (
	"orglang/orglang/adt/expctx"
)

type SigSpecME struct {
	X     expctx.BindClaimME   `json:"x"`
	SigQN string               `json:"sig_qn"`
	Ys    []expctx.BindClaimME `json:"ys"`
}

type IdentME struct {
	SigID string `json:"id" param:"id"`
}

type SigRefME struct {
	SigID string `json:"id" param:"id"`
	Title string `json:"title"`
	SigRN int64  `json:"rev"`
}

type SigSnapME struct {
	X     expctx.BindClaimME   `json:"x"`
	SigID string               `json:"sig_id"`
	Ys    []expctx.BindClaimME `json:"ys"`
	Title string               `json:"title"`
	SigRN int64                `json:"sig_rn"`
}
