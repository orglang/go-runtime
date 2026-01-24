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
	MsgToExecSpec   func(poolexec.ExecSpec) (ExecSpec, error)
	MsgFromExecSpec func(ExecSpec) poolexec.ExecSpec
	MsgToExecRef    func(poolexec.ExecRef) (ExecRef, error)
	MsgFromExecRef  func(ExecRef) poolexec.ExecRef
	MsgToExecSnap   func(poolexec.ExecSnap) (ExecSnap, error)
	MsgFromExecSnap func(ExecSnap) poolexec.ExecSnap
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
)
