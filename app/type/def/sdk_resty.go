package def

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

func (cl *clientResty) Incept(roleQN sym.ADT) (TypeRef, error) {
	return TypeRef{}, nil
}

func (cl *clientResty) Create(spec TypeSpec) (TypeSnap, error) {
	req := MsgFromTypeSpec(spec)
	var res TypeSnapMsg
	resp, err := cl.resty.R().
		SetResult(&res).
		SetBody(&req).
		Post("/roles")
	if err != nil {
		return TypeSnap{}, err
	}
	if resp.IsError() {
		return TypeSnap{}, fmt.Errorf("received: %v", string(resp.Body()))
	}
	return MsgToTypeSnap(res)
}

func (c *clientResty) Modify(snap TypeSnap) (TypeSnap, error) {
	return TypeSnap{}, nil
}

func (c *clientResty) Retrieve(rid id.ADT) (TypeSnap, error) {
	return TypeSnap{}, nil
}

func (c *clientResty) retrieveSnap(entity TypeRec) (TypeSnap, error) {
	return TypeSnap{}, nil
}

func (c *clientResty) RetreiveRefs() ([]TypeRef, error) {
	return []TypeRef{}, nil
}
