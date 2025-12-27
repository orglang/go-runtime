package dec

import (
	"fmt"

	"github.com/go-resty/resty/v2"

	"orglang/orglang/avt/id"
	"orglang/orglang/avt/sym"
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

func (cl *clientResty) Incept(sigQN sym.ADT) (ProcRef, error) {
	return ProcRef{}, nil
}

func (cl *clientResty) Create(spec ProcSpec) (ProcSnap, error) {
	req := MsgFromSigSpec(spec)
	var res SigSnapMsg
	resp, err := cl.resty.R().
		SetResult(&res).
		SetBody(&req).
		Post("/signatures")
	if err != nil {
		return ProcSnap{}, err
	}
	if resp.IsError() {
		return ProcSnap{}, fmt.Errorf("received: %v", string(resp.Body()))
	}
	return MsgToSigSnap(res)
}

func (c *clientResty) Retrieve(id id.ADT) (ProcSnap, error) {
	var res SigSnapMsg
	resp, err := c.resty.R().
		SetResult(&res).
		SetPathParam("id", id.String()).
		Get("/signatures/{id}")
	if err != nil {
		return ProcSnap{}, err
	}
	if resp.IsError() {
		return ProcSnap{}, fmt.Errorf("received: %v", string(resp.Body()))
	}
	return MsgToSigSnap(res)
}

func (c *clientResty) RetreiveRefs() ([]ProcRef, error) {
	refs := []ProcRef{}
	return refs, nil
}
