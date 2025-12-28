package dec

type SigSpecVP struct {
	SigNS string `form:"ns" json:"ns"`
	SigSN string `form:"sn" json:"sn"`
}

type SigRefVP struct {
	SigID string `form:"sig_id" json:"sig_id" param:"id"`
	SigRN int64  `form:"sig_rn" json:"sig_rn"`
	Title string `form:"name" json:"title"`
}

type SigSnapVP struct {
	SigID string `json:"sig_id"`
	SigRN int64  `json:"sig_rn"`
	Title string `json:"title"`
}
