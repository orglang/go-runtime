package procexec

import (
	"orglang/orglang/adt/procexp"
)

type ExecSpecME struct {
	ExecID string            `json:"exec_id" param:"id"`
	PoolID string            `json:"pool_id"`
	ProcES procexp.ExpSpecME `json:"term"`
}

type IdentME struct {
	ExecID string `param:"id"`
}

type ExecRefME struct {
	ExecID string `json:"exec_id"`
}

type ExecSnapME struct {
	ExecID string `json:"exec_id"`
}
