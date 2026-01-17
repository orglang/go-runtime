package poolexec

import (
	"orglang/go-runtime/adt/procexec"

	"github.com/orglang/go-sdk/adt/poolexec"
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
var (
	ConvertRecToRef func(ExecRec) ExecRef
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
// goverter:extend orglang/go-runtime/adt/procdef:Msg.*
var (
	MsgToExecSpec   func(poolexec.ExecSpecME) (ExecSpec, error)
	MsgFromExecSpec func(ExecSpec) poolexec.ExecSpecME
	MsgToExecRef    func(poolexec.ExecRefME) (ExecRef, error)
	MsgFromExecRef  func(ExecRef) poolexec.ExecRefME
	MsgToExecSnap   func(poolexec.ExecSnapME) (ExecSnap, error)
	MsgFromExecSnap func(ExecSnap) poolexec.ExecSnapME
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
var (
	DataToExecRef    func(execRefDS) (ExecRef, error)
	DataFromExecRef  func(ExecRef) execRefDS
	DataToExecRefs   func([]execRefDS) ([]ExecRef, error)
	DataFromExecRefs func([]ExecRef) []execRefDS
	DataToExecRec    func(execRecDS) (ExecRec, error)
	DataFromExecRec  func(ExecRec) execRecDS
	DataToLiab       func(liabDS) (procexec.Liab, error)
	DataFromLiab     func(procexec.Liab) liabDS
	DataToExecSnap   func(execSnapDS) (ExecSnap, error)
	DataFromExecSnap func(ExecSnap) execSnapDS
	DataToEPs        func([]epDS) ([]procexec.EP, error)
)
