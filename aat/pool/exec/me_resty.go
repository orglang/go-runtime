package exec

import (
	"github.com/go-resty/resty/v2"

	"orglang/orglang/avt/id"

	procexec "orglang/orglang/aat/proc/exec"
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

func (cl *sdkResty) Create(spec PoolSpec) (PoolRef, error) {
	req := MsgFromPoolSpec(spec)
	var res PoolRefME
	_, err := cl.resty.R().
		SetResult(&res).
		SetBody(&req).
		Post("/pools")
	if err != nil {
		return PoolRef{}, err
	}
	return MsgToPoolRef(res)
}

func (cl *sdkResty) Poll(spec PollSpec) (procexec.ProcRef, error) {
	return procexec.ProcRef{}, nil
}

func (cl *sdkResty) Retrieve(poolID id.ADT) (PoolSnap, error) {
	var res PoolSnapME
	_, err := cl.resty.R().
		SetResult(&res).
		SetPathParam("id", poolID.String()).
		Get("/pools/{id}")
	if err != nil {
		return PoolSnap{}, err
	}
	return MsgToPoolSnap(res)
}

func (cl *sdkResty) RetreiveRefs() ([]PoolRef, error) {
	refs := []PoolRef{}
	return refs, nil
}

func (cl *sdkResty) Spawn(spec procexec.ProcSpec) (procexec.ProcRef, error) {
	req := procexec.MsgFromSpec(spec)
	var res procexec.RefME
	_, err := cl.resty.R().
		SetResult(&res).
		SetBody(&req).
		SetPathParam("poolID", spec.PoolID.String()).
		Post("/pools/{poolID}/procs")
	if err != nil {
		return procexec.ProcRef{}, err
	}
	return procexec.MsgToRef(res)
}

func (cl *sdkResty) Take(spec StepSpec) error {
	req := MsgFromStepSpec(spec)
	var res procexec.RefME
	_, err := cl.resty.R().
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
