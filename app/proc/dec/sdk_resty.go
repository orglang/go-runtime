package dec

import (
	"fmt"

	"github.com/go-resty/resty/v2"

	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"
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

func (cl *clientResty) Incept(sigQN sym.ADT) (SigRef, error) {
	return SigRef{}, nil
}

func (cl *clientResty) Create(spec SigSpec) (SigSnap, error) {
	req := MsgFromSigSpec(spec)
	var res SigSnapMsg
	resp, err := cl.resty.R().
		SetResult(&res).
		SetBody(&req).
		Post("/signatures")
	if err != nil {
		return SigSnap{}, err
	}
	if resp.IsError() {
		return SigSnap{}, fmt.Errorf("received: %v", string(resp.Body()))
	}
	return MsgToSigSnap(res)
}

func (c *clientResty) Retrieve(id id.ADT) (SigSnap, error) {
	var res SigSnapMsg
	resp, err := c.resty.R().
		SetResult(&res).
		SetPathParam("id", id.String()).
		Get("/signatures/{id}")
	if err != nil {
		return SigSnap{}, err
	}
	if resp.IsError() {
		return SigSnap{}, fmt.Errorf("received: %v", string(resp.Body()))
	}
	return MsgToSigSnap(res)
}

func (c *clientResty) RetreiveRefs() ([]SigRef, error) {
	refs := []SigRef{}
	return refs, nil
}
