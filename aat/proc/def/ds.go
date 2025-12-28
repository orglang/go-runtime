package def

import (
	"orglang/orglang/lib/sd"
)

type Repo interface {
	InsertProc(sd.Source, ProcRec) error
}

type TermRecDS struct {
	K     termKind    `json:"k"`
	Close *closeRecDS `json:"close,omitempty"`
	Wait  *waitRecDS  `json:"wait,omitempty"`
	Send  *sendRecDS  `json:"send,omitempty"`
	Recv  *recvRecDS  `json:"recv,omitempty"`
	Lab   *labRecDS   `json:"lab,omitempty"`
	Case  *caseRecDS  `json:"case,omitempty"`
	Fwd   *fwdRecDS   `json:"fwd,omitempty"`
}

type TermSpecDS struct {
	K     termKind     `json:"k"`
	Close *closeSpecDS `json:"close,omitempty"`
	Wait  *waitSpecDS  `json:"wait,omitempty"`
	Send  *sendSpecDS  `json:"send,omitempty"`
	Recv  *recvSpecDS  `json:"recv,omitempty"`
	Lab   *labSpecDS   `json:"lab,omitempty"`
	Case  *caseSpecDS  `json:"case,omitempty"`
	Fwd   *fwdSpecDS   `json:"fwd,omitempty"`
}

type termKind int

const (
	nonterm = termKind(iota)
	closeKind
	waitKind
	sendKind
	recvKind
	labKind
	caseKind
	linkKind
	spawnKind
	fwdKind
)

type closeSpecDS struct {
	X string `json:"x"`
}

type closeRecDS struct {
	X string `json:"x"`
}

type waitSpecDS struct {
	X    string     `json:"x"`
	Cont TermSpecDS `json:"cont"`
}

type waitRecDS struct {
	X    string     `json:"x"`
	Cont TermSpecDS `json:"cont"`
}

type sendSpecDS struct {
	X string `json:"x"`
	Y string `json:"y"`
}

type sendRecDS struct {
	X string `json:"x"`
	A string `json:"a"`
	B string `json:"b"`
}

type recvSpecDS struct {
	X    string     `json:"x"`
	Y    string     `json:"y"`
	Cont TermSpecDS `json:"cont"`
}

type recvRecDS struct {
	X    string     `json:"x"`
	A    string     `json:"a"`
	Y    string     `json:"y"`
	Cont TermSpecDS `json:"cont"`
}

type labSpecDS struct {
	X     string `json:"x"`
	Label string `json:"lab"`
}

type labRecDS struct {
	X     string `json:"x"`
	Label string `json:"lab"`
}

type caseSpecDS struct {
	X        string         `json:"x"`
	Branches []branchSpecDS `json:"brs"`
}

type caseRecDS struct {
	X        string        `json:"x"`
	Branches []branchRecDS `json:"brs"`
}

type branchSpecDS struct {
	Label string     `json:"lab"`
	Cont  TermSpecDS `json:"cont"`
}

type branchRecDS struct {
	Label string     `json:"lab"`
	Cont  TermSpecDS `json:"cont"`
}

type fwdSpecDS struct {
	X string `json:"x"`
	Y string `json:"y"`
}

type fwdRecDS struct {
	X string `json:"x"`
	B string `json:"b"`
}
