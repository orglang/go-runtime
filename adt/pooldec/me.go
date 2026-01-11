package pooldec

import (
	"orglang/orglang/adt/termctx"
)

type DecSpecME struct {
	PoolNS               string
	PoolSN               string
	InsiderProvisionBC   termctx.BindClaimME
	InsiderReceptionBC   termctx.BindClaimME
	OutsiderProvisionBCs []termctx.BindClaimME
	OutsiderReceptionBCs []termctx.BindClaimME
}

type DecRefME struct{}
