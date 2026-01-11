package poolstep

import (
	"orglang/orglang/adt/procexp"
)

type StepSpecME struct {
	ExecID string            `json:"exec_id"`
	ProcID string            `json:"proc_id"`
	Term   procexp.ExpSpecME `json:"term"`
}
