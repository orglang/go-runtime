package procbind

type BindSpecDS struct {
	ChnlPH string `json:"chnl_ph"`
	TypeQN string `json:"type_qn"`
}

type BindRecDS struct {
	ID     string `db:"exec_id"`
	RN     int64  `db:"exec_rn"`
	ChnlBS uint8  `db:"chnl_bs"`
	ChnlPH string `db:"chnl_ph"`
	ChnlID string `db:"chnl_id"`
	ExpID  string `db:"exp_id"`
}
