package procexp

type expKindME string

const (
	Close = expKindME("close")
	Wait  = expKindME("wait")
	Send  = expKindME("send")
	Recv  = expKindME("recv")
	Lab   = expKindME("lab")
	Case  = expKindME("case")
	Call  = expKindME("call")
	Link  = expKindME("link")
	Spawn = expKindME("spawn")
	Fwd   = expKindME("fwd")
)

type ExpSpecME struct {
	K       expKindME      `json:"kind"`
	Close   *CloseSpecME   `json:"close,omitempty"`
	Wait    *WaitSpecME    `json:"wait,omitempty"`
	Send    *SendSpecME    `json:"send,omitempty"`
	Recv    *RecvSpecME    `json:"recv,omitempty"`
	Lab     *LabSpecME     `json:"lab,omitempty"`
	Case    *CaseSpecME    `json:"case,omitempty"`
	Spawn   *SpawnSpecME   `json:"spawn,omitempty"`
	Fwd     *FwdSpecME     `json:"fwd,omitempty"`
	Call    *CallSpecME    `json:"call,omitempty"`
	Accept  *AcceptSpecME  `json:"accept,omitempty"`
	Detach  *DetachSpecME  `json:"detach,omitempty"`
	Acquire *AcqureSpecME  `json:"acquire,omitempty"`
	Release *ReleaseSpecME `json:"release,omitempty"`
}

type CloseSpecME struct {
	CommPH string `json:"comm_ph"`
}

type WaitSpecME struct {
	CommPH string    `json:"comm_ph"`
	ContES ExpSpecME `json:"cont_es"`
}

type SendSpecME struct {
	CommPH string `json:"comm_ph"`
	ValPH  string `json:"val_ph"`
}

type RecvSpecME struct {
	CommPH string    `json:"comm_ph"`
	BindPH string    `json:"bind_ph"`
	ContES ExpSpecME `json:"cont_es"`
}

type LabSpecME struct {
	CommPH string `json:"comm_ph"`
	Label  string `json:"label"`
}

type CaseSpecME struct {
	CommPH  string         `json:"comm_ph"`
	ContBSs []BranchSpecME `json:"cont_bss"`
}

type BranchSpecME struct {
	Label  string    `json:"label"`
	ContES ExpSpecME `json:"cont_es"`
}

type CallSpecME struct {
	CommPH string    `json:"comm_ph"`
	BindPH string    `json:"bind_ph"`
	ProcQN string    `json:"proc_qn"` // раньше был SigPH
	ValPHs []string  `json:"val_phs"`
	ContES ExpSpecME `json:"cont_es"`
}

type SpawnSpecME struct {
	CommPH  string     `json:"comm_ph"`
	DecID   string     `json:"dec_id"`
	ProcQN  string     `json:"proc_qn"`
	BindPHs []string   `json:"bind_phs"`
	ContES  *ExpSpecME `json:"cont_es"`
}

type FwdSpecME struct {
	X string `json:"x"`
	Y string `json:"y"`
}

type AcceptSpecME struct {
	CommPH string    `json:"comm_ph"`
	ContES ExpSpecME `json:"cont_es"`
}

type DetachSpecME struct {
	CommPH string `json:"comm_ph"`
}

type AcqureSpecME struct {
	CommPH string    `json:"comm_ph"`
	ContES ExpSpecME `json:"cont_es"`
}

type ReleaseSpecME struct {
	CommPH string `json:"comm_ph"`
}
