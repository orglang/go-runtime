package exec

import (
	"github.com/go-resty/resty/v2"

	"orglang/orglang/avt/id"
)

// Client-side secondary adapter
type sdkResty struct {
	resty *resty.Client
}

func newSdkResty() *sdkResty {
	r := resty.New().SetBaseURL("http://localhost:8080/api/v1")
	return &sdkResty{r}
}

func NewAPI() API {
	return newSdkResty()
}

func (cl *sdkResty) Run(spec ProcSpec) error {
	req := MsgFromSpec(spec)
	var res RefME
	_, err := cl.resty.R().
		SetPathParam("id", spec.ExecID.String()).
		SetBody(&req).
		SetResult(&res).
		Post("/procs/{id}/calls")
	if err != nil {
		return err
	}
	return nil
}

func (cl *sdkResty) Retrieve(procID id.ADT) (ProcSnap, error) {
	var res SnapME
	_, err := cl.resty.R().
		SetPathParam("id", procID.String()).
		SetResult(&res).
		Get("/procs/{id}")
	if err != nil {
		return ProcSnap{}, err
	}
	return MsgToSnap(res)
}
