package root

import (
	"github.com/go-resty/resty/v2"

	"smecalculus/rolevod/lib/id"
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
		SetPathParam("id", spec.ProcID.String()).
		SetBody(&req).
		SetResult(&res).
		Post("/procs/{id}/calls")
	if err != nil {
		return Ref{}, err
	}
	return MsgToRef(res)
}

func (cl *clientResty) Retrieve(procID id.ADT) (Snap, error) {
	var res SnapMsg
	_, err := cl.resty.R().
		SetPathParam("id", procID.String()).
		SetResult(&res).
		Get("/procs/{id}")
	if err != nil {
		return Snap{}, err
	}
	return MsgToSnap(res)
}
