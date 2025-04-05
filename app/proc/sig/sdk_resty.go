package sig

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

func (cl *clientResty) Incept(sigQN sym.ADT) (Ref, error) {
	return Ref{}, nil
}

func (cl *clientResty) Create(spec Spec) (Impl, error) {
	req := MsgFromSpec(spec)
	var res ImplMsg
	resp, err := cl.resty.R().
		SetResult(&res).
		SetBody(&req).
		Post("/signatures")
	if err != nil {
		return Impl{}, err
	}
	if resp.IsError() {
		return Impl{}, fmt.Errorf("received: %v", string(resp.Body()))
	}
	return MsgToRoot(res)
}

func (c *clientResty) Retrieve(id id.ADT) (Impl, error) {
	var res ImplMsg
	resp, err := c.resty.R().
		SetResult(&res).
		SetPathParam("id", id.String()).
		Get("/signatures/{id}")
	if err != nil {
		return Impl{}, err
	}
	if resp.IsError() {
		return Impl{}, fmt.Errorf("received: %v", string(resp.Body()))
	}
	return MsgToRoot(res)
}

func (c *clientResty) RetreiveRefs() ([]Ref, error) {
	refs := []Ref{}
	return refs, nil
}
