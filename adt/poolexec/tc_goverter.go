package poolexec

import (
	"github.com/orglang/go-sdk/adt/poolexec"
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
// goverter:extend orglang/go-runtime/adt/uniqsym:Convert.*
// goverter:extend orglang/go-runtime/adt/uniqref:Msg.*
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
// goverter:extend orglang/go-runtime/adt/uniqref:Data.*
var (
	DataToExecRefs   func([]execRefDS) ([]ExecRef, error)
	DataFromExecRefs func([]ExecRef) []execRefDS
	// goverter:map . ExecRef
	DataToExecRec func(execRecDS) (ExecRec, error)
	// goverter:autoMap ExecRef
	DataFromExecRec func(ExecRec) execRecDS
	// goverter:map . ExecRef
	DataToLiab func(liabDS) (Liab, error)
	// goverter:autoMap ExecRef
	DataFromLiab func(Liab) liabDS
	// goverter:map . ExecRef
	DataToExecSnap func(execSnapDS) (ExecSnap, error)
	// goverter:autoMap ExecRef
	DataFromExecSnap func(ExecSnap) execSnapDS
)
