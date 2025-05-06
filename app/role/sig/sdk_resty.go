package role

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

func (cl *clientResty) Incept(roleQN sym.ADT) (Ref, error) {
	return Ref{}, nil
}

func (cl *clientResty) Create(spec Spec) (Snap, error) {
	req := MsgFromSpec(spec)
	var res SnapMsg
	resp, err := cl.resty.R().
		SetResult(&res).
		SetBody(&req).
		Post("/roles")
	if err != nil {
		return Snap{}, err
	}
	if resp.IsError() {
		return Snap{}, fmt.Errorf("received: %v", string(resp.Body()))
	}
	return MsgToSnap(res)
}

func (c *clientResty) Modify(snap Snap) (Snap, error) {
	return Snap{}, nil
}

func (c *clientResty) Retrieve(rid id.ADT) (Snap, error) {
	return Snap{}, nil
}

func (c *clientResty) RetrieveImpl(rid id.ADT) (Impl, error) {
	var res RootMsg
	resp, err := c.resty.R().
		SetResult(&res).
		SetPathParam("id", rid.String()).
		Get("/roles/{id}")
	if err != nil {
		return Impl{}, err
	}
	if resp.IsError() {
		return Impl{}, fmt.Errorf("received: %v", string(resp.Body()))
	}
	return MsgToRoot(res)
}

func (c *clientResty) RetrieveSnap(entity Impl) (Snap, error) {
	return Snap{}, nil
}

func (c *clientResty) RetreiveRefs() ([]Ref, error) {
	return []Ref{}, nil
}
