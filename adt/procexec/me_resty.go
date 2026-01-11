package procexec

import (
	"github.com/go-resty/resty/v2"

	"orglang/orglang/adt/identity"
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
		Post("/procs/{id}/calls")
	if err != nil {
		return err
	}
	return nil
}

func (sdk *RestySDK) Retrieve(procID identity.ADT) (ExecSnapME, error) {
	var res ExecSnapME
	_, err := sdk.Client.R().
		SetPathParam("id", procID.String()).
		SetResult(&res).
		Get("/procs/{id}")
	if err != nil {
		return ExecSnapME{}, err
	}
	return res, nil
}
