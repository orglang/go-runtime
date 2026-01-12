package procexec

import (
	"github.com/go-resty/resty/v2"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/procstep"
)

// Client-side secondary adapter
type RestySDK struct {
	Client *resty.Client
}

func (sdk *RestySDK) Run(spec ExecSpecME) error {
	var res ExecRefME
	_, err := sdk.Client.R().
		SetPathParam("id", spec.ExecID).
		SetBody(&spec).
		SetResult(&res).
		Post("/procs/{id}/execs")
	if err != nil {
		return err
	}
	return nil
}

func (sdk *RestySDK) Take(spec procstep.StepSpec) error {
	req := procstep.MsgFromStepSpec(spec)
	var res ExecRefME
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

func (sdk *RestySDK) Retrieve(execID identity.ADT) (ExecSnapME, error) {
	var res ExecSnapME
	_, err := sdk.Client.R().
		SetPathParam("id", execID.String()).
		SetResult(&res).
		Get("/procs/{id}")
	if err != nil {
		return ExecSnapME{}, err
	}
	return res, nil
}
