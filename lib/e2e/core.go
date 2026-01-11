package e2e

import (
	"github.com/go-resty/resty/v2"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/pooldec"
	"orglang/orglang/adt/poolexec"
	"orglang/orglang/adt/procdec"
	"orglang/orglang/adt/procexec"
	"orglang/orglang/adt/typedef"
)

type PoolDecAPI interface {
	Create(pooldec.DecSpecME) (pooldec.DecRefME, error)
}

func newPoolDecAPI(client *resty.Client) PoolDecAPI {
	return &pooldec.RestySDK{Client: client}
}

type PoolExecAPI interface {
	Retrieve(identity.ADT) (poolexec.ExecSnap, error)
	Create(poolexec.ExecSpec) (poolexec.ExecRef, error)
	Poll(poolexec.PollSpec) (procexec.ExecRef, error)
}

func newPoolExecAPI(client *resty.Client) PoolExecAPI {
	return &poolexec.RestySDK{Client: client}
}

type ProcDecAPI interface {
	Create(procdec.DecSpec) (procdec.DecSnap, error)
}

func newProcDecAPI(client *resty.Client) ProcDecAPI {
	return &procdec.RestySDK{Client: client}
}

type ProcExecAPI interface {
	Run(procexec.ExecSpecME) error
}

func newProcExecAPI(client *resty.Client) ProcExecAPI {
	return &procexec.RestySDK{Client: client}
}

type TypeDefAPI interface {
	Create(typedef.DefSpec) (typedef.DefSnap, error)
}

func newTypeDefAPI(client *resty.Client) TypeDefAPI {
	return &typedef.RestySDK{Client: client}
}
