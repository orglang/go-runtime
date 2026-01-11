package poolexec

type ExecSpecME struct {
	PoolQN  string   `json:"pool_qn"`
	ProcIDs []string `json:"proc_ids"`
	SupID   string   `json:"sup_id"`
}

type IdentME struct {
	PoolID string `json:"id" param:"id"`
}

type ExecRefME struct {
	ExecID string `json:"exec_id"`
	ProcID string `json:"proc_id"`
}

type ExecSnapME struct {
	ExecID string      `json:"exec_id"`
	Title  string      `json:"title"`
	Subs   []ExecRefME `json:"subs"`
}
