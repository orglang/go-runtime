package def

type TypeSpecVP struct {
	NS   string `form:"ns" json:"ns"`
	Name string `form:"name" json:"name"`
}

type TypeRefVP struct {
	RoleID string `form:"id" json:"id" param:"id"`
	RoleRN int64  `form:"rev" json:"rev"`
	Title  string `form:"name" json:"title"`
}

type TypeSnapVP struct {
	RoleID string     `json:"id"`
	RoleRN int64      `json:"rev"`
	Title  string     `json:"title"`
	State  TermSpecME `json:"state"`
}
