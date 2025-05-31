package exec

import (
	"github.com/go-resty/resty/v2"

	"smecalculus/rolevod/lib/id"

	procexec "smecalculus/rolevod/app/proc/exec"
)

// Adapter
type clientResty struct {
	resty *resty.Client
}

func newClientResty() *clientResty {
	r := resty.New().SetBaseURL("http://localhost:8080/api/v1")
	return &clientResty{r}
}

func NewAPI() API {
	return newClientResty()
}

func (cl *clientResty) Create(spec PoolSpec) (PoolRef, error) {
	req := MsgFromPoolSpec(spec)
	var res PoolRefMsg
	_, err := cl.resty.R().
		SetResult(&res).
		SetBody(&req).
		Post("/pools")
	if err != nil {
		return PoolRef{}, err
	}
	return MsgToPoolRef(res)
}

func (cl *clientResty) Poll(spec PollSpec) (procexec.ProcRef, error) {
	return procexec.ProcRef{}, nil
}

func (cl *clientResty) Retrieve(poolID id.ADT) (PoolSnap, error) {
	var res PoolSnapMsg
	_, err := cl.resty.R().
		SetResult(&res).
		SetPathParam("id", poolID.String()).
		Get("/pools/{id}")
	if err != nil {
		return PoolSnap{}, err
	}
	return MsgToPoolSnap(res)
}

func (cl *clientResty) RetreiveRefs() ([]PoolRef, error) {
	refs := []PoolRef{}
	return refs, nil
}

func (cl *clientResty) Spawn(spec procexec.ProcSpec) (procexec.ProcRef, error) {
	req := procexec.MsgFromSpec(spec)
	var res procexec.RefMsg
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

func (cl *clientResty) Take(spec StepSpec) error {
	req := MsgFromStepSpec(spec)
	var res procexec.RefMsg
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
