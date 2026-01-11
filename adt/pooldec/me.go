package pooldec

import (
	"orglang/orglang/adt/termctx"
)

type DecSpecME struct {
	PoolNS               string
	PoolSN               string
	InsiderProvisionEP   termctx.BindClaimME
	InsiderReceptionEP   termctx.BindClaimME
	OutsiderProvisionEP  termctx.BindClaimME
	OutsiderReceptionEPs []termctx.BindClaimME
}

type DecRefME struct{}
