package dec

import (
	"fmt"

	"github.com/go-resty/resty/v2"

	"orglang/orglang/avt/id"
	"orglang/orglang/avt/sym"
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

func (cl *sdkResty) Incept(sigQN sym.ADT) (ProcRef, error) {
	return ProcRef{}, nil
}

func (cl *sdkResty) Create(spec ProcSpec) (ProcSnap, error) {
	req := MsgFromSigSpec(spec)
	var res SigSnapME
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

func (c *sdkResty) Retrieve(id id.ADT) (ProcSnap, error) {
	var res SigSnapME
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

func (c *sdkResty) RetreiveRefs() ([]ProcRef, error) {
	refs := []ProcRef{}
	return refs, nil
}
