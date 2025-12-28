package def

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

func (cl *sdkResty) Incept(roleQN sym.ADT) (TypeRef, error) {
	return TypeRef{}, nil
}

func (cl *sdkResty) Create(spec TypeSpec) (TypeSnap, error) {
	req := MsgFromTypeSpec(spec)
	var res TypeSnapME
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

func (c *sdkResty) Modify(snap TypeSnap) (TypeSnap, error) {
	return TypeSnap{}, nil
}

func (c *sdkResty) Retrieve(rid id.ADT) (TypeSnap, error) {
	return TypeSnap{}, nil
}

func (c *sdkResty) retrieveSnap(entity TypeRec) (TypeSnap, error) {
	return TypeSnap{}, nil
}

func (c *sdkResty) RetreiveRefs() ([]TypeRef, error) {
	return []TypeRef{}, nil
}
