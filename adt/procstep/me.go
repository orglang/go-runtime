package procstep

import (
	"orglang/orglang/adt/procexp"
)

type StepSpecME struct {
	ExecID string            `json:"exec_id"`
	ProcID string            `json:"proc_id"`
	ProcES procexp.ExpSpecME `json:"proc_es"`
}
