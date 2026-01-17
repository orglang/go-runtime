package e2e

import (
	"github.com/go-resty/resty/v2"

	"github.com/orglang/go-sdk/adt/pooldec"
	"github.com/orglang/go-sdk/adt/poolexec"
	"github.com/orglang/go-sdk/adt/procdec"
	"github.com/orglang/go-sdk/adt/procexec"
	"github.com/orglang/go-sdk/adt/typedef"
)

type PoolDecAPI interface {
	Create(pooldec.DecSpecME) (pooldec.DecRefME, error)
}

func newPoolDecAPI(client *resty.Client) PoolDecAPI {
	return &pooldec.RestySDK{Client: client}
}

type PoolExecAPI interface {
	Retrieve(string) (poolexec.ExecSnapME, error)
	Create(poolexec.ExecSpecME) (poolexec.ExecRefME, error)
	Poll(poolexec.PollSpecME) (procexec.ExecRefME, error)
}

func newPoolExecAPI(client *resty.Client) PoolExecAPI {
	return &poolexec.RestySDK{Client: client}
}

type ProcDecAPI interface {
	Create(procdec.DecSpecME) (procdec.DecSnapME, error)
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
	Create(typedef.DefSpecME) (typedef.DefSnapME, error)
}

func newTypeDefAPI(client *resty.Client) TypeDefAPI {
	return &typedef.RestySDK{Client: client}
}
