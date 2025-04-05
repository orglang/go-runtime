package pool

import (
	"github.com/go-resty/resty/v2"

	"smecalculus/rolevod/lib/id"

	procroot "smecalculus/rolevod/app/proc/root"
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

func (cl *clientResty) Create(spec Spec) (Ref, error) {
	req := MsgFromSpec(spec)
	var res RefMsg
	_, err := cl.resty.R().
		SetResult(&res).
		SetBody(&req).
		Post("/pools")
	if err != nil {
		return Ref{}, err
	}
	return MsgToRef(res)
}

func (cl *clientResty) Retrieve(rid id.ADT) (Snap, error) {
	var res SnapMsg
	_, err := cl.resty.R().
		SetResult(&res).
		SetPathParam("id", rid.String()).
		Get("/pools/{id}")
	if err != nil {
		return Snap{}, err
	}
	return MsgToSnap(res)
}

func (cl *clientResty) RetreiveRefs() ([]Ref, error) {
	refs := []Ref{}
	return refs, nil
}

func (cl *clientResty) Spawn(spec procroot.Spec) (procroot.Ref, error) {
	req := procroot.MsgFromSpec(spec)
	var res procroot.RefMsg
	_, err := cl.resty.R().
		SetResult(&res).
		SetBody(&req).
		SetPathParam("poolID", spec.PoolID.String()).
		Post("/pools/{poolID}/procs")
	if err != nil {
		return procroot.Ref{}, err
	}
	return procroot.MsgToRef(res)
}

func (cl *clientResty) Take(spec StepSpec) error {
	req := MsgFromStepSpec(spec)
	var res procroot.RefMsg
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
