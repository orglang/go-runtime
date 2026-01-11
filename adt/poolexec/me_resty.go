package poolexec

import (
	"github.com/go-resty/resty/v2"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/procexec"
)

// Client-side secondary adapter
type RestySDK struct {
	Client *resty.Client
}

func (sdk *RestySDK) Create(spec ExecSpec) (ExecRef, error) {
	req := MsgFromExecSpec(spec)
	var res ExecRefME
	_, err := sdk.Client.R().
		SetResult(&res).
		SetBody(&req).
		Post("/pools")
	if err != nil {
		return ExecRef{}, err
	}
	return MsgToExecRef(res)
}

func (sdk *RestySDK) Poll(spec PollSpec) (procexec.ExecRef, error) {
	return procexec.ExecRef{}, nil
}

func (sdk *RestySDK) Retrieve(poolID identity.ADT) (ExecSnap, error) {
	var res ExecSnapME
	_, err := sdk.Client.R().
		SetResult(&res).
		SetPathParam("id", poolID.String()).
		Get("/pools/{id}")
	if err != nil {
		return ExecSnap{}, err
	}
	return MsgToExecSnap(res)
}

func (sdk *RestySDK) RetreiveRefs() ([]ExecRef, error) {
	refs := []ExecRef{}
	return refs, nil
}

func (sdk *RestySDK) Spawn(spec procexec.ExecSpec) (procexec.ExecRef, error) {
	req := procexec.MsgFromExecSpec(spec)
	var res procexec.ExecRefME
	_, err := sdk.Client.R().
		SetResult(&res).
		SetBody(&req).
		SetPathParam("poolID", spec.PoolID.String()).
		Post("/pools/{poolID}/procs")
	if err != nil {
		return procexec.ExecRef{}, err
	}
	return procexec.MsgToExecRef(res)
}

func (sdk *RestySDK) Take(spec StepSpec) error {
	req := MsgFromStepSpec(spec)
	var res procexec.ExecRefME
	_, err := sdk.Client.R().
		SetResult(&res).
		SetBody(&req).
		SetPathParam("poolID", spec.PoolID.String()).
		SetPathParam("procID", spec.ProcID.String()).
		Post("/pools/{poolID}/procs/{procID}/steps")
	if err != nil {
		return err
	}
	return nil
}
