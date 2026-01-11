package poolstep

import (
	"github.com/go-resty/resty/v2"

	"orglang/orglang/adt/procexec"
)

// Client-side secondary adapter
type RestySDK struct {
	Client *resty.Client
}

func (sdk *RestySDK) Poll(spec PollSpec) (procexec.ExecRef, error) {
	return procexec.ExecRef{}, nil
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
		SetPathParam("poolID", spec.ExecID.String()).
		SetPathParam("procID", spec.ProcID.String()).
		Post("/pools/{poolID}/procs/{procID}/steps")
	if err != nil {
		return err
	}
	return nil
}
